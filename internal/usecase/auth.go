package usecase

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	"github.com/Weit145/gateway-startup/internal/domain/health"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/google/uuid"
)

func (s *Service) CheckHealthAuth(ctx context.Context) (*health.Health, error) {

	authStatus := "success"
	dbHealth := "success"
	var healthErr error

	authHealthy, grpcErr := s.grpcClient.CheckHealth(ctx)
	if grpcErr != nil {
		authStatus = "error"
		dbHealth = "error"
		healthErr = logger.WrapError(ctx, fmt.Errorf("check auth health: %w: %w", grpcErr, errs.ErrUnavailable))
	} else if !authHealthy {
		authStatus = "error"
		dbHealth = "error"
		healthErr = logger.WrapError(ctx, fmt.Errorf("auth service reported unhealthy: %w", errs.ErrUnavailable))
	}

	authHealth, err := health.NewHealth(ctx, "auth", authStatus, &dbHealth)
	if err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("check health auth grpc: %w", err))
	}

	return &authHealth, healthErr
}

func (s *Service) Register(ctx context.Context, email string, password string, name string) error {
	if err := s.redis.LimiterEmail(ctx, email); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("rate limit email: %w", err))
	}
	if err := s.grpcClient.Register(ctx, email, password, name); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("register user: %w", err))
	}
	return nil
}

func (s *Service) RepeatEmail(ctx context.Context, email string) error {
	if err := s.redis.LimiterEmail(ctx, email); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("rate limit email: %w", err))
	}
	if err := s.grpcClient.RepeatEmail(ctx, email); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("reepat email register: %w", err))
	}
	return nil
}

func (s *Service) ConfirmEmail(ctx context.Context, token string, userID uuid.UUID) (auth.User, auth.Tokens, error) {
	user, tokens, err := s.grpcClient.ConfirmEmail(ctx, token, userID.String())
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("confirm email: %w", err))
	}
	return user, tokens, nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (auth.User, auth.Tokens, error) {
	user, tokens, err := s.grpcClient.Login(ctx, email, password)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("login user: %w", err))
	}

	return user, tokens, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	err := s.grpcClient.Logout(ctx, userID.String(), refreshToken)
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("logout user: %w", err))
	}
	return nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (auth.User, auth.Tokens, error) {
	user, tokens, err := s.grpcClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		return auth.User{}, auth.Tokens{}, logger.WrapError(ctx, fmt.Errorf("refresh token: %w", err))
	}
	return user, tokens, nil
}

func (s *Service) PasswordRequest(ctx context.Context, email string) error {
	if err := s.redis.LimiterEmail(ctx, email); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("rate limit email: %w", err))
	}
	if err := s.grpcClient.PasswordRequest(ctx, email); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("password email request: %w", err))
	}
	return nil
}

func (s *Service) PasswordConfirm(ctx context.Context, token string, userID uuid.UUID, newPassword string) error {
	if err := s.grpcClient.PasswordConfirm(ctx, token, userID, newPassword); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("password confirm: %w", err))
	}
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error {
	if err := s.grpcClient.ChangePassword(ctx, userID, oldPassword, newPassword); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("change password: %w", err))
	}
	return nil
}
