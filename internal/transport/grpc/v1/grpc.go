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
	CheckHealth(ctx context.Context) (bool, error)
	Register(ctx context.Context, email string, password string, name string) error
	Login(ctx context.Context, email string, password string) (auth.User, auth.Tokens, error)
	ConfirmEmail(ctx context.Context, token string, userID string) (auth.User, auth.Tokens, error)
	Logout(ctx context.Context, userID string, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (auth.User, auth.Tokens, error)
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

func (c *Client) Register(ctx context.Context, email string, password string, name string) error {
	_, err := c.api.Register(ctx, &GRPCauth.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	})

	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to register via gRPC", "error", err)
		return logger.WrapError(ctx, err)
	}

	return nil
}

func (c *Client) Login(ctx context.Context, email string, password string) (auth.User, auth.Tokens, error) {
	resp, err := c.api.Login(ctx, &GRPCauth.LoginRequest{
		Email:    email,
		Password: password,
	})

	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to login via gRPC", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, err)
	}

	accessToken := resp.GetAccessToken()
	refreshToken := resp.GetRefreshToken()
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
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, err)
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		slog.DebugContext(ctx, "Received invalid user ID from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	user, err := auth.New(id, email, username, role)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create user from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	tokens, err := auth.NewTokens(accessToken, refreshToken)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create tokens from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	return user, tokens, nil
}

func (c *Client) ConfirmEmail(ctx context.Context, token string, userID string) (auth.User, auth.Tokens, error) {
	resp, err := c.api.ConfirmEmail(ctx, &GRPCauth.ConfirmRequest{
		Token:  token,
		UserID: userID,
	})

	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to confirm email via gRPC", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, err)
	}

	accessToken := resp.GetAccessToken()
	refreshToken := resp.GetRefreshToken()
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
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, err)
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		slog.DebugContext(ctx, "Received invalid user ID from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	user, err := auth.New(id, email, username, role)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create user from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}
	tokens, err := auth.NewTokens(accessToken, refreshToken)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create tokens from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	return user, tokens, nil
}

func (c *Client) Logout(ctx context.Context, userID string, refreshToken string) error {
	_, err := c.api.Logout(ctx, &GRPCauth.LogOutRequest{
		UserID:       userID,
		RefreshToken: refreshToken,
	})

	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to Logout via gRPC", "error", err)
		return logger.WrapError(ctx, err)
	}

	return nil
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (auth.User, auth.Tokens, error) {
	resp, err := c.api.RefreshToken(ctx, &GRPCauth.RefreshRequest{
		RefreshToken: refreshToken,
	})

	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to refresh token via gRPC", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, err)
	}

	accessToken := resp.GetAccessToken()
	responseRefreshToken := resp.GetRefreshToken()
	idString := resp.GetUser().GetId()
	email := resp.GetUser().GetEmail()
	role := resp.GetUser().GetRole()
	username := resp.GetUser().GetUsername()

	err = mapper_err_grpc.InvalidErr(ctx,
		mapper_err_grpc.Field{Name: "accessToken", Value: accessToken},
		mapper_err_grpc.Field{Name: "refreshToken", Value: responseRefreshToken},
		mapper_err_grpc.Field{Name: "id", Value: idString},
		mapper_err_grpc.Field{Name: "email", Value: email},
		mapper_err_grpc.Field{Name: "username", Value: username},
		mapper_err_grpc.Field{Name: "role", Value: role},
	)

	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, err)
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		slog.DebugContext(ctx, "Received invalid user ID from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	user, err := auth.New(id, email, username, role)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create user from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	tokens, err := auth.NewTokens(accessToken, responseRefreshToken)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create tokens from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	return user, tokens, nil
}
