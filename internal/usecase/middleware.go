package usecase

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/google/uuid"
)

func (s *Service) CheckLogin(ctx context.Context, authHeader string) (uuid.UUID, uuid.UUID, error) {
	userID, jwtId, err := s.jwt.CheckToken(ctx, authHeader)
	if err != nil {
		return uuid.Nil, uuid.Nil, logger.WrapError(ctx, fmt.Errorf("failed to check bearer token: %w", err))
	}

	return userID, jwtId, nil
}

func (s *Service) CheckIPRateLimit(ctx context.Context, ip string) error {
	if err := s.redis.LimiterIP(ctx, ip); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("rate limit: %w", err))
	}
	return nil
}

func (s *Service) CheckUserRateLimit(ctx context.Context, userID uuid.UUID) error {
	if err := s.redis.LimiterUserID(ctx, userID); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("rate limit: %w", err))
	}
	return nil
}

func (s *Service) CheckCompleteRegistration(ctx context.Context, authHeader string) (uuid.UUID, uuid.UUID, error) {
	session, jwtID, err := s.jwt.CheckToken(ctx, authHeader)
	if err != nil {
		return uuid.Nil, uuid.Nil, logger.WrapError(ctx, fmt.Errorf("failed to check bearer token: %w", err))
	}
	return session, jwtID, nil
}
