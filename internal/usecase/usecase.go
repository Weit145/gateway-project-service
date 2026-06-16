package usecase

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-startup/internal/domain/health"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/usecase/port"
)

type Service struct {
	grpcClient port.GRPServices
	redis      port.RedisServices
	jwt        port.JWTServices
}

func NewServices(grpcClient port.GRPServices, redis port.RedisServices, jwt port.JWTServices) *Service {
	return &Service{
		grpcClient: grpcClient,
		redis:      redis,
		jwt:        jwt,
	}
}

func (s *Service) CheckHealthGateway(ctx context.Context) (*health.Health, error) {
	gatewayStatus := "success"
	redisStatus := "success"
	var healthErr error

	if err := s.redis.CheckHealth(ctx); err != nil {
		gatewayStatus = "error"
		redisStatus = "error"
		healthErr = logger.WrapError(ctx, fmt.Errorf("check redis health: %w: %w", err, errs.ErrUnavailable))
	}

	gatewayHealth, err := health.NewHealth(ctx, "gateway", gatewayStatus, &redisStatus)
	if err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	return &gatewayHealth, healthErr
}
