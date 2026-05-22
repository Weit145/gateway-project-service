package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	mapper_err_grpc "github.com/Weit145/gateway-startup/internal/transport/grpc/v1/err"
	"github.com/google/uuid"

	GRPCauth "github.com/weit145/proto-startup/auth"
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
	Register(ctx context.Context, email string, password string, name string) (err error, health bool)
	Login(ctx context.Context, email string, password string) (err error, accessToken string, refreshToken string, user auth.User)
	ConfirmEmail(ctx context.Context, token string) (err error, accessToken string, refreshToken string, user auth.User)
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
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to check health via gRPC", "error", err)
		return logger.WrapError(ctx, err), false
	}

	health = resp.GetHealth()
	if health == false {
		slog.DebugContext(ctx, "Health check returned false from gRPC response")
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), false
	}

	return nil, health
}

func (c *Client) Register(ctx context.Context, email string, password string, name string) (err error, health bool) {
	resp, err := c.api.Register(ctx, &GRPCauth.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	})

	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to register via gRPC", "error", err)
		return logger.WrapError(ctx, err), false
	}

	health = resp.GetHealth()
	if health == false {
		slog.DebugContext(ctx, "Health check returned false from gRPC response")
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), false
	}

	return nil, health
}

func (c *Client) Login(ctx context.Context, email string, password string) (err error, accessToken string, refreshToken string, user auth.User) {
	resp, err := c.api.Login(ctx, &GRPCauth.LoginRequest{
		Email:    email,
		Password: password,
	})

	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to login via gRPC", "error", err)
		return logger.WrapError(ctx, err), "", "", auth.User{}
	}

	accessToken = resp.GetAccessToken()
	refreshToken = resp.GetRefreshToken()
	idString := resp.GetUser().GetId()
	email = resp.GetUser().GetEmail()
	role := resp.GetUser().GetRole()
	username := resp.GetUser().GetUsername()

	err = mapper_err_grpc.InvalidErr(ctx,
		mapper_err_grpc.Field{Name: "accessToken", Value: accessToken},
		mapper_err_grpc.Field{Name: "refreshToken", Value: refreshToken},
		mapper_err_grpc.Field{Name: "id", Value: idString},
		mapper_err_grpc.Field{Name: "email", Value: email},
		mapper_err_grpc.Field{Name: "username", Value: username},
		mapper_err_grpc.Field{Name: "role", Value: role},
	)

	if err != nil {
		return logger.WrapError(ctx, err), "", "", auth.User{}
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		slog.DebugContext(ctx, "Received invalid user ID from gRPC response", "error", err)
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}

	user, err = auth.New(id, email, username, role)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create user from gRPC response", "error", err)
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}

	return nil, accessToken, refreshToken, user
}

func (c *Client) ConfirmEmail(ctx context.Context, token string) (err error, accessToken string, refreshToken string, user auth.User) {
	resp, err := c.api.ConfirmEmail(ctx, &GRPCauth.ConfirmRequest{
		Token: token,
	})

	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to confirm email via gRPC", "error", err)
		return logger.WrapError(ctx, err), "", "", auth.User{}
	}

	accessToken = resp.GetAccessToken()
	refreshToken = resp.GetRefreshToken()
	idString := resp.GetUser().GetId()
	email := resp.GetUser().GetEmail()
	role := resp.GetUser().GetRole()
	username := resp.GetUser().GetUsername()

	err = mapper_err_grpc.InvalidErr(ctx,
		mapper_err_grpc.Field{Name: "accessToken", Value: accessToken},
		mapper_err_grpc.Field{Name: "refreshToken", Value: refreshToken},
		mapper_err_grpc.Field{Name: "id", Value: idString},
		mapper_err_grpc.Field{Name: "email", Value: email},
		mapper_err_grpc.Field{Name: "username", Value: username},
		mapper_err_grpc.Field{Name: "role", Value: role},
	)

	if err != nil {
		return logger.WrapError(ctx, err), "", "", auth.User{}
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		slog.DebugContext(ctx, "Received invalid user ID from gRPC response", "error", err)
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}

	user, err = auth.New(id, email, username, role)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create user from gRPC response", "error", err)
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal)), "", "", auth.User{}
	}

	return nil, accessToken, refreshToken, user
}
