package grpc

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-project-service/internal/logger"
	mapper_err_grpc "github.com/Weit145/gateway-project-service/internal/transport/grpc/v1/err"
	"github.com/google/uuid"
	GRPCauth "github.com/weit145/proto-startup/auth"
)

func (c *Client) PasswordRequest(ctx context.Context, email string) error {
	rpcCtx, cancel := c.rpcLongContext(ctx)
	defer cancel()

	if _, err := c.api.PasswordRequest(rpcCtx, &GRPCauth.Email{Email: email}); err != nil {
		err = mapper_err_grpc.MapErr(err)
		return logger.WrapError(ctx, fmt.Errorf("From gRPC: %w", err))
	}
	return nil
}

func (c *Client) PasswordConfirm(ctx context.Context, token string, userID uuid.UUID, newPassword string) error {
	rpcCtx, cancel := c.rpcDefaultContext(ctx)
	defer cancel()

	if _, err := c.api.PasswordConfirm(rpcCtx, &GRPCauth.PasswordConfirmRequest{
		Token:       token,
		UserID:      userID.String(),
		NewPassword: newPassword,
	}); err != nil {
		err = mapper_err_grpc.MapErr(err)
		return logger.WrapError(ctx, fmt.Errorf("From gRPC: %w", err))
	}
	return nil
}

func (c *Client) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error {
	rpcCtx, cancel := c.rpcLongContext(ctx)
	defer cancel()

	if _, err := c.api.ChangePassword(rpcCtx, &GRPCauth.ChangePasswordRequest{
		UserID:      userID.String(),
		NewPassword: newPassword,
		OldPassword: oldPassword,
	}); err != nil {
		err = mapper_err_grpc.MapErr(err)
		return logger.WrapError(ctx, fmt.Errorf("From gRPC: %w", err))
	}
	return nil
}
