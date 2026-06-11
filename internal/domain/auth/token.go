package auth

import (
	"fmt"

	"github.com/Weit145/gateway-startup/internal/lib/valid"
)

type Tokens struct {
	accessToken  string
	refreshToken string
}

func NewTokens(accessToken string, refreshToken string) (Tokens, error) {
	accessToken, err := valid.ValidateOpaqueValue(accessToken)
	if err != nil {
		return Tokens{}, fmt.Errorf("invalid access token: %w", err)
	}

	refreshToken, err = valid.ValidateOpaqueValue(refreshToken)
	if err != nil {
		return Tokens{}, fmt.Errorf("invalid refresh token: %w", err)
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
