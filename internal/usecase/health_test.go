package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/google/uuid"
)

type healthGRPCStub struct {
	healthy bool
	err     error
}

func (s healthGRPCStub) CheckHealth(context.Context) (bool, error) {
	return s.healthy, s.err
}

func (healthGRPCStub) Register(context.Context, string, string, string) error {
	return nil
}

func (healthGRPCStub) Login(context.Context, string, string) (auth.User, auth.Tokens, error) {
	return auth.User{}, auth.Tokens{}, nil
}

func (healthGRPCStub) ConfirmEmail(context.Context, string, string) (auth.User, auth.Tokens, error) {
	return auth.User{}, auth.Tokens{}, nil
}

func (healthGRPCStub) Logout(context.Context, string, string) error {
	return nil
}

func (healthGRPCStub) RefreshToken(context.Context, string) (auth.User, auth.Tokens, error) {
	return auth.User{}, auth.Tokens{}, nil
}

type healthRedisStub struct {
	err error
}

func (s healthRedisStub) CheckHealth(context.Context) error {
	return s.err
}

func (healthRedisStub) LimiterIP(context.Context, string) error {
	return nil
}

func (healthRedisStub) LimiterUserID(context.Context, uuid.UUID) error {
	return nil
}

func (healthRedisStub) LimiterEmail(context.Context, string) error {
	return nil
}

type healthJWTStub struct{}

func (healthJWTStub) CheckToken(string) (uuid.UUID, uuid.UUID, error) {
	return uuid.Nil, uuid.Nil, nil
}

func TestCheckHealthGatewayReportsRedisFailure(t *testing.T) {
	service := NewServices(healthGRPCStub{healthy: true}, healthRedisStub{err: errors.New("redis unavailable")}, healthJWTStub{})

	got, err := service.CheckHealthGateway(context.Background())

	if !errors.Is(err, errs.ErrUnavailable) {
		t.Fatalf("CheckHealthGateway() error = %v, want ErrUnavailable", err)
	}
	if got == nil || got.Status() != "error" {
		t.Fatalf("gateway health = %#v, want error status", got)
	}
	if got.DB() == nil || *got.DB() != "error" {
		t.Fatalf("gateway DB health = %v, want error", got.DB())
	}
}

func TestCheckHealthAuthReportsDependencyFailure(t *testing.T) {
	tests := []struct {
		name string
		grpc healthGRPCStub
	}{
		{name: "RPC error", grpc: healthGRPCStub{err: errors.New("auth unavailable")}},
		{name: "unhealthy response", grpc: healthGRPCStub{healthy: false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewServices(tt.grpc, healthRedisStub{}, healthJWTStub{})

			got, err := service.CheckHealthAuth(context.Background())

			if !errors.Is(err, errs.ErrUnavailable) {
				t.Fatalf("CheckHealthAuth() error = %v, want ErrUnavailable", err)
			}
			if got == nil || got.Status() != "error" {
				t.Fatalf("auth health = %#v, want error status", got)
			}
			if got.DB() == nil || *got.DB() != "error" {
				t.Fatalf("auth DB health = %v, want error", got.DB())
			}
		})
	}
}
