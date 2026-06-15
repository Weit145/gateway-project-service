package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Weit145/gateway-startup/docs"
	"github.com/Weit145/gateway-startup/internal/config"
	"github.com/Weit145/gateway-startup/internal/logger"
	my_redis "github.com/Weit145/gateway-startup/internal/repositories/storage/redis"
	v1 "github.com/Weit145/gateway-startup/internal/transport/api/v1"
	"github.com/Weit145/gateway-startup/internal/transport/grpc/v1"
	"github.com/Weit145/gateway-startup/internal/usecase"
)

// @title Gateway Startup API
// @version 1.0
// @description HTTP gateway for startup services.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter: Bearer <access_token>
func main() {
	if err := run(); err != nil {
		slog.ErrorContext(logger.ErrorCtx(context.Background(), err), "gateway stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	//Init config
	cfg, err := config.InitConfig()
	if err != nil {
		return fmt.Errorf("init config: %w", err)
	}

	//Init logger
	logger.InitLogging(cfg.Log.Level)
	slog.Info("Init logger")

	//Init context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//Init grpc client
	grpcClient, err := grpc.New(cfg.GRPC.AuthAddr, grpc.TLSSettings{
		CertPEM:    cfg.GRPC.TLSCertPEM,
		KeyPEM:     cfg.GRPC.TLSKeyPEM,
		ServerCA:   cfg.GRPC.TLSServerCAPEM,
		ServerName: cfg.GRPC.TLSServerName,
	}, grpc.TimeSettings{
		Short:   cfg.GRPC.TimeoutShort,
		Default: cfg.GRPC.TimeoutDefault,
		Long:    cfg.GRPC.TimeoutLong,
	})
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("init grpc client: %w", err))
	}
	slog.Info("Init grpc")
	defer grpcClient.Close()

	//Init redis
	redis, err := my_redis.NewRepository(ctx, my_redis.Setting{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		Protocol: cfg.Redis.Protocol,
	})
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("init redis client: %w", err))
	}
	slog.Info("Init redis")
	defer redis.Close()

	//Init usecase
	ser := usecase.NewServices(grpcClient, redis)
	slog.Info("Init usecase")

	//Init api
	router := v1.NewRouter(ser, redis, cfg.HTTP.TrustedProxies, cfg.JWT.PublicKey)
	slog.Info("Init api")

	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("http server started", "addr", cfg.HTTP.Address)

	if err := serveHTTP(ctx, server, 10*time.Second); err != nil {
		return logger.WrapError(ctx, err)
	}
	return nil
}

type httpServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

func serveHTTP(ctx context.Context, server httpServer, shutdownTimeout time.Duration) error {
	shutdownSignal, stopShutdown := context.WithCancel(ctx)
	defer stopShutdown()

	shutdownDone := make(chan error, 1)
	go func() {
		<-shutdownSignal.Done()
		if ctx.Err() == nil {
			shutdownDone <- nil
			return
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		shutdownDone <- server.Shutdown(shutdownCtx)
	}()

	serveErr := server.ListenAndServe()
	if ctx.Err() != nil {
		if err := <-shutdownDone; err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
	}
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", serveErr)
	}
	return nil
}
