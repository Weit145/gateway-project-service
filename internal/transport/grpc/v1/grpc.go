package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"

	GRPCauth "github.com/Weit145/proto-startup/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api  GRPCauth.AuthClient
	conn *grpc.ClientConn
}

type GRPServices interface {
	CheckHealth(ctx context.Context) (err error, health bool)
	Register(ctx context.Context, email string, password string) (err error, accessToken string, refreshToken string, user auth.User)
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

func (c *Client) CheckHealth(ctx context.Context) (err error, health bool) {
	resp, err := c.api.CheckHealth(ctx, &GRPCauth.Empty{})
	if err != nil {
		slog.DebugContext(ctx, "Failed to check health via gRPC", "error", err)
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), false
	}

	health = resp.GetHealth()
	if health == false {
		slog.DebugContext(ctx, "Health check returned false from gRPC response")
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), false
	}

	return nil, health
}

func (c *Client) Register(ctx context.Context, email string, password string) (err error, accessToken string, refreshToken string, user auth.User) {
	resp, err := c.api.Register(ctx, &GRPCauth.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		slog.DebugContext(ctx, "Failed to register via gRPC", "error", err)
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}

	accessToken = resp.GetAccessToken()
	if accessToken == "" {
		slog.DebugContext(ctx, "Received empty access token from gRPC response")
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}
	refreshToken = resp.GetRefreshToken()
	if refreshToken == "" {
		slog.DebugContext(ctx, "Received empty refresh token from gRPC response")
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}

	idString := resp.GetUser().GetId()
	if idString == "" {
		slog.DebugContext(ctx, "Received empty user ID from gRPC response")
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}
	id, err := uuid.Parse(idString)
	if err != nil {
		slog.DebugContext(ctx, "Received invalid user ID from gRPC response", "error", err)
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}
	email = resp.GetUser().GetEmail()
	if email == "" {
		slog.DebugContext(ctx, "Received empty user email from gRPC response")
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}
	role := resp.GetUser().GetRole()
	if role == "" {
		slog.DebugContext(ctx, "Received empty user role from gRPC response")
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}
	username := resp.GetUser().GetUsername()
	if username == "" {
		slog.DebugContext(ctx, "Received empty user username from gRPC response")
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}

	user, err = auth.New(id, email, username, role)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create user from gRPC response", "error", err)
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}

	return nil, accessToken, refreshToken, user
}
