package usecase

import (
	"context"

	"github.com/Weit145/gateway-startup/internal/transport/grpc/v1"
)

type Service struct {
	grpcClient grpc.GRPServices
}

type Services interface {
	CheckHealth(ctx context.Context) error
}

func NewServices(grpcClient grpc.GRPServices) *Service {
	return &Service{
		grpcClient: grpcClient,
	}
}
