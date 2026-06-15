package handler_auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	"github.com/Weit145/gateway-startup/internal/domain/health"
	"github.com/google/uuid"
)

type refreshServiceStub struct {
	user   auth.User
	tokens auth.Tokens
}

func (s refreshServiceStub) CheckHealthGateway(context.Context) (*health.Health, error) {
	return nil, nil
}

func (s refreshServiceStub) CheckHealthAuth(context.Context) (*health.Health, error) {
	return nil, nil
}

func (s refreshServiceStub) Register(context.Context, string, string, string) error {
	return nil
}

func (s refreshServiceStub) Login(context.Context, string, string) (auth.User, auth.Tokens, error) {
	return auth.User{}, auth.Tokens{}, nil
}

func (s refreshServiceStub) ConfirmEmail(context.Context, string, uuid.UUID) (auth.User, auth.Tokens, error) {
	return auth.User{}, auth.Tokens{}, nil
}

func (s refreshServiceStub) Logout(context.Context, uuid.UUID, string) error {
	return nil
}

func (s refreshServiceStub) RefreshToken(context.Context, string) (auth.User, auth.Tokens, error) {
	return s.user, s.tokens, nil
}

func TestRefreshTokenReturnsOK(t *testing.T) {
	userID := uuid.New()
	user, err := auth.NewUser(userID, "user@example.com", "username", "user")
	if err != nil {
		t.Fatalf("auth.NewUser() error = %v", err)
	}
	tokens, err := auth.NewTokens("access-token", "new-refresh-token")
	if err != nil {
		t.Fatalf("auth.NewTokens() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh-token"})
	rec := httptest.NewRecorder()

	RefreshToken(refreshServiceStub{user: user, tokens: tokens}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
