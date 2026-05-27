package usecase

import (
	"context"
	"fmt"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
	"github.com/google/uuid"
)

type AuthService interface {
	CheckHealth(ctx context.Context) error
	Register(ctx context.Context, email string, password string, name string) error
	Login(ctx context.Context, email string, password string) (error, dto.RegisterResponse, string)
	ConfirmEmail(ctx context.Context, token string, userID uuid.UUID) (error, dto.RegisterResponse, string)
}

func (s *Service) CheckHealth(ctx context.Context) error {
	err, health := s.grpcClient.CheckHealth(ctx)
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}
	if !health {
		return logger.WrapError(ctx, fmt.Errorf("%w", errs.ErrInternal))
	}
	return nil
}

func (s *Service) Register(ctx context.Context, email string, password string, name string) error {
	err, health := s.grpcClient.Register(ctx, email, password, name)
	if err != nil {
		return logger.WrapError(ctx, err)
	}
	if !health {
		return logger.WrapError(ctx, errs.ErrInternal)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (error, dto.RegisterResponse, string) {
	err, accessToken, refreshToken, user := s.grpcClient.Login(ctx, email, password)
	if err != nil {
		return logger.WrapError(ctx, err), dto.RegisterResponse{}, ""
	}
	return nil, dto.RegisterResponse{
		AccessToken: accessToken,
		User: dto.User{
			Id:    user.ID(),
			Email: user.Email(),
			Name:  user.Username(),
			Role:  user.Role(),
		},
	}, refreshToken
}

func (s *Service) ConfirmEmail(ctx context.Context, token string, userID uuid.UUID) (error, dto.RegisterResponse, string) {
	err, accessToken, refreshToken, user := s.grpcClient.ConfirmEmail(ctx, token, userID.String())
	if err != nil {
		return logger.WrapError(ctx, err), dto.RegisterResponse{}, ""
	}
	return nil, dto.RegisterResponse{
		AccessToken: accessToken,
		User: dto.User{
			Id:    user.ID(),
			Email: user.Email(),
			Name:  user.Username(),
			Role:  user.Role(),
		},
	}, refreshToken
}
