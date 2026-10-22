package grpc

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-project-service/internal/domain/auth"
	errs "github.com/Weit145/gateway-project-service/internal/lib/e"
	"github.com/Weit145/gateway-project-service/internal/logger"
	mapper_err_grpc "github.com/Weit145/gateway-project-service/internal/transport/grpc/v1/err"
	"github.com/google/uuid"
	GRPCauth "github.com/weit145/proto-startup/auth"
)

func (c *Client) Register(ctx context.Context, email string) error {
	rpcCtx, cancel := c.rpcLongContext(ctx)
	defer cancel()

	_, err := c.api.Register(rpcCtx, &GRPCauth.RegisterRequest{
		Email: email,
	})
	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		return logger.WrapError(ctx, fmt.Errorf("From gRPC: %w", err))
	}
	return nil
}

func (c *Client) ConfirmEmail(ctx context.Context, token string, userID string) (auth.Token, error) {
	rpcCtx, cancel := c.rpcDefaultContext(ctx)
	defer cancel()

	resp, err := c.api.ConfirmEmail(rpcCtx, &GRPCauth.ConfirmRequest{
		Token:  token,
		UserID: userID,
	})
	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		return auth.Token{}, logger.WrapError(ctx, fmt.Errorf("From gRPC: %w", err))
	}

	accessToken := resp.GetAccessToken()

	err = mapper_err_grpc.InvalidErr(ctx,
		mapper_err_grpc.Field{Name: "accessToken", Value: accessToken},
	)
	if err != nil {
		return auth.Token{}, logger.WrapError(ctx, fmt.Errorf("Invalid value from gRPC response: %w", err))
	}

	tokenDomain, err := auth.NewToken(accessToken)
	if err != nil {
		return auth.Token{}, logger.WrapError(ctx, fmt.Errorf("Faild create domain tokens: %w", err))
	}

	return tokenDomain, nil
}

func (c *Client) RegisterComplete(ctx context.Context, password string, session uuid.UUID) (auth.User, auth.Tokens, error) {
	rpcCtx, cancel := c.rpcDefaultContext(ctx)
	defer cancel()

	resp, err := c.api.RegisterComplete(rpcCtx, &GRPCauth.RegisterCompleteRequest{
		Password: password,
		Session:  session.String(),
	})
	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("From gRPC: %w", err))
	}

	accessToken := resp.GetAccessToken()
	refreshToken := resp.GetRefreshToken()
	idString := resp.GetUser().GetId()
	email := resp.GetUser().GetEmail()
	role := resp.GetUser().GetRole()

	err = mapper_err_grpc.InvalidErr(ctx,
		mapper_err_grpc.Field{Name: "accessToken", Value: accessToken},
		mapper_err_grpc.Field{Name: "refreshToken", Value: refreshToken},
		mapper_err_grpc.Field{Name: "id", Value: idString},
		mapper_err_grpc.Field{Name: "email", Value: email},
		mapper_err_grpc.Field{Name: "role", Value: role},
	)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Invalid value from gRPC response: %w", err))
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Invalid user ID from gRPC response: %w", errs.ErrInternal))
	}

	user, err := auth.NewUser(id, email, role)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Faild create domain user: %w", err))
	}

	accessTokenDomain, err := auth.NewToken(accessToken)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Faild create domain tokens: %w", err))
	}
	refreshTokenDomain, err := auth.NewToken(refreshToken)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Faild create domain tokens: %w", err))
	}
	tokens, err := auth.NewTokens(accessTokenDomain, refreshTokenDomain)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("Faild create domain tokens: %w", err))
	}

	return user, tokens, nil
}

func (c *Client) RepeatEmail(ctx context.Context, email string) error {
	rpcCtx, cancel := c.rpcDefaultContext(ctx)
	defer cancel()

	if _, err := c.api.RepeatEmail(rpcCtx, &GRPCauth.Email{Email: email}); err != nil {
		err = mapper_err_grpc.MapErr(err)
		return logger.WrapError(ctx, fmt.Errorf("From gRPC: %w", err))
	}
	return nil
}
