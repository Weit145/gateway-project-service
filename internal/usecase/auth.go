package usecase

import (
	"context"
	"fmt"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
)

type AuthService interface {
}

func (s *Service) CheckHealth(ctx context.Context) error {
	result, err := s.grpcClient.CheckHealth(ctx)
	if err != nil {
		return fmt.Errorf("%w", errs.ErrInternal)
	}
	if result.GetHealth() != true {
		return fmt.Errorf("%w", errs.ErrInternal)
	}
	return nil
}
