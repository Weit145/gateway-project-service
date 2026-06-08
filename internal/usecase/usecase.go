package usecase

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-startup/internal/domain/health"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/grpc/v1"
)

type Service struct {
	grpcClient grpc.GRPServices
}

type Services interface {
	CheckHealthGateway(ctx context.Context) (*health.Health, error)
	AuthService
}

func NewServices(grpcClient grpc.GRPServices) *Service {
	return &Service{
		grpcClient: grpcClient,
	}
}

func (s *Service) CheckHealthGateway(ctx context.Context) (*health.Health, error) {
	gatewayHealth, err := health.NewHealth(ctx, "gateway", "success", nil)
	if err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}

	return &gatewayHealth, nil
}
