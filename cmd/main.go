package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Weit145/gateway-startup/docs"
	"github.com/Weit145/gateway-startup/internal/config"
	"github.com/Weit145/gateway-startup/internal/logger"
	v1 "github.com/Weit145/gateway-startup/internal/transport/api/v1"
	"github.com/Weit145/gateway-startup/internal/transport/grpc/v1"
	"github.com/Weit145/gateway-startup/internal/usecase"
)

// @title Gateway Startup API
// @version 1.0
// @description HTTP gateway for startup services.
// @host localhost:8080
// @BasePath /
func main() {

	//Init config
	cfg := config.InitConfig()

	//Init logger
	logger.InitLogging(cfg.Log.Level)
	slog.Info("Init logger")

	//Init context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//Init grpc client
	grpcClient, err := grpc.New(cfg.GRPC.AuthAddr)
	if err != nil {
		slog.ErrorContext(logger.ErrorCtx(ctx, err), "init grpc client", "error", err)
		return
	}
	slog.Info("Init grpc")
	defer grpcClient.Close()

	//Init usecase
	ser := usecase.NewServices(grpcClient)

	//Init api
	router := v1.NewRouter(ser)

	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.ErrorContext(logger.ErrorCtx(shutdownCtx, err), "shutdown http server", "error", err)
		}
	}()
	slog.Info("http server started", "addr", cfg.HTTP.Address)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.ErrorContext(logger.ErrorCtx(ctx, err), "listen and serve", "error", err)
		os.Exit(1)
	}
}
