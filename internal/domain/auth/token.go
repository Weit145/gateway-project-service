package auth

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-startup/internal/lib/valid"
	"github.com/Weit145/gateway-startup/internal/logger"
)

type Tokens struct {
	accessToken  string
	refreshToken string
}

func NewTokens(ctx context.Context, accessToken string, refreshToken string) (Tokens, error) {
	accessToken, err := valid.ValidateOpaqueValue(ctx, accessToken)
	if err != nil {
		return Tokens{}, logger.WrapError(ctx, fmt.Errorf("invalid access token: %w", err))
	}

	refreshToken, err = valid.ValidateOpaqueValue(ctx, refreshToken)
	if err != nil {
		return Tokens{}, logger.WrapError(ctx, fmt.Errorf("invalid refresh token: %w", err))
	}

	return Tokens{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

func (t Tokens) AccessToken() string {
	return t.accessToken
}

func (t Tokens) RefreshToken() string {
	return t.refreshToken
}
