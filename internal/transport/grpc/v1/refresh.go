package grpc

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	mapper_err_grpc "github.com/Weit145/gateway-startup/internal/transport/grpc/v1/err"
	"github.com/google/uuid"

	GRPCauth "github.com/weit145/proto-startup/auth"
)

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (auth.User, auth.Tokens, error) {
	rpcCtx, cancel := c.rpcDefaultContext(ctx)
	defer cancel()

	resp, err := c.api.RefreshToken(rpcCtx, &GRPCauth.RefreshRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("From gRPC: %w", err))
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
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Invalid value from gRPC response: %w", err))
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Invalid user ID from gRPC response: %w", errs.ErrInternal))
	}

	user, err := auth.NewUser(id, email, username, role)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Faild create domain user: %w", err))
	}

	tokens, err := auth.NewTokens(accessToken, responseRefreshToken)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Faild create domain tokens: %w", err))
	}

	return user, tokens, nil
}
