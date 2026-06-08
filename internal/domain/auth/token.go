package auth

import (
	"fmt"
	"strings"
)

type Tokens struct {
	accessToken  string
	refreshToken string
}

func NewTokens(accessToken string, refreshToken string) (Tokens, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return Tokens{}, fmt.Errorf("invalid access token")
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return Tokens{}, fmt.Errorf("invalid refresh token")
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
