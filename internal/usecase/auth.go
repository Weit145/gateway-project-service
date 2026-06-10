package usecase

import (
	"context"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	"github.com/Weit145/gateway-startup/internal/domain/health"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/google/uuid"
)

type AuthService interface {
	CheckHealthAuth(ctx context.Context) (*health.Health, error)
	Register(ctx context.Context, email string, password string, name string) error
	Login(ctx context.Context, email string, password string) (auth.User, auth.Tokens, error)
	ConfirmEmail(ctx context.Context, token string, userID uuid.UUID) (auth.User, auth.Tokens, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (auth.User, auth.Tokens, error)
}

func (s *Service) CheckHealthAuth(ctx context.Context) (*health.Health, error) {

	authStatus := "success"
	dbHealth := "success"
	authHealthy, grpcErr := s.grpcClient.CheckHealth(ctx)
	if grpcErr != nil {
		authStatus = "error"
		dbHealth = "error"
	}
	if !authHealthy {
		dbHealth = "error"
	}

	authHealth, err := health.NewHealth(ctx, "auth", authStatus, &dbHealth)
	if err != nil {
		return nil, logger.WrapError(ctx, errs.ErrInternal)
	}

	return &authHealth, nil
}

func (s *Service) Register(ctx context.Context, email string, password string, name string) error {
	err := s.grpcClient.Register(ctx, email, password, name)
	if err != nil {
		return logger.WrapError(ctx, err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (auth.User, auth.Tokens, error) {
	user, tokens, err := s.grpcClient.Login(ctx, email, password)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, err)
	}

	return user, tokens, nil
}

func (s *Service) ConfirmEmail(ctx context.Context, token string, userID uuid.UUID) (auth.User, auth.Tokens, error) {
	user, tokens, err := s.grpcClient.ConfirmEmail(ctx, token, userID.String())
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, err)
	}
	return user, tokens, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	err := s.grpcClient.Logout(ctx, userID.String(), refreshToken)
	if err != nil {
		return logger.WrapError(ctx, err)
	}
	return nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (auth.User, auth.Tokens, error) {
	user, tokens, err := s.grpcClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, err)
	}
	return user, tokens, nil
}
