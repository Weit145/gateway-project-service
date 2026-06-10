package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	mapper_err_grpc "github.com/Weit145/gateway-startup/internal/transport/grpc/v1/err"

	GRPCauth "github.com/weit145/proto-startup/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api  GRPCauth.AuthClient
	conn *grpc.ClientConn
}

func New(addr string) (*Client, error) {
	if addr == "" {
		slog.Error("Empty address provided for gRPC client")
		return nil, fmt.Errorf("%w", errs.ErrInternal)
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to create gRPC client connection", "error", err)
		return nil, fmt.Errorf("%w", errs.ErrInternal)
	}

	conn.Connect()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}

		if !conn.WaitForStateChange(ctx, state) {
			slog.Error("gRPC connection did not become ready", "state", state)
			return nil, fmt.Errorf("%w", errs.ErrInternal)
		}
	}

	grpcClient := GRPCauth.NewAuthClient(conn)

	return &Client{
		api:  grpcClient,
		conn: conn,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) CheckHealth(ctx context.Context) (bool, error) {
	resp, err := c.api.CheckHealth(ctx, &GRPCauth.Empty{})
	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to check health via gRPC", "error", err)
		return false, logger.WrapError(ctx, err)
	}

	return resp.GetHealth(), nil
}
