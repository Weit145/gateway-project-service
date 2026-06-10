package grpc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	mapper_err_grpc "github.com/Weit145/gateway-startup/internal/transport/grpc/v1/err"
	"github.com/google/uuid"

	GRPCauth "github.com/weit145/proto-startup/auth"
)

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

	user, err := auth.NewUser(ctx, id, email, username, role)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create user from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	tokens, err := auth.NewTokens(ctx, accessToken, responseRefreshToken)
	if err != nil {
		slog.DebugContext(ctx, "Failed to create tokens from gRPC response", "error", err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	return user, tokens, nil
}
