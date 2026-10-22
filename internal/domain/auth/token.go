package auth

import (
	"fmt"

	"github.com/Weit145/gateway-project-service/internal/lib/valid"
)

type Token struct {
	token string
}

func NewToken(token string) (Token, error) {
	token, err := valid.ValidateOpaqueValue(token)
	if err != nil {
		return Token{}, fmt.Errorf("invalid token: %w", err)
	}
	return Token{token: token}, nil
}

func (t Token) Token() string { return t.token }

type Tokens struct {
	accessToken  Token
	refreshToken Token
}

// TODO: STRING OR TOKEN STRUCT
func NewTokens(accessToken Token, refreshToken Token) (Tokens, error) {
	return Tokens{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

func (t Tokens) AccessToken() Token  { return t.accessToken }
func (t Tokens) RefreshToken() Token { return t.refreshToken }
