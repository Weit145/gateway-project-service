package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Weit145/gateway-startup/internal/domain/health"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
)

type healthServiceStub struct {
	gateway    *health.Health
	gatewayErr error
	auth       *health.Health
	authErr    error
}

func (s healthServiceStub) CheckHealthGateway(context.Context) (*health.Health, error) {
	return s.gateway, s.gatewayErr
}

func (s healthServiceStub) CheckHealthAuth(context.Context) (*health.Health, error) {
	return s.auth, s.authErr
}

func TestHealthReturnsServiceUnavailableWithDiagnostics(t *testing.T) {
	ctx := context.Background()
	redisStatus := "error"
	authDBStatus := "success"
	gateway, err := health.NewHealth(ctx, "gateway", "error", &redisStatus)
	if err != nil {
		t.Fatalf("health.NewHealth(gateway) error = %v", err)
	}
	authHealth, err := health.NewHealth(ctx, "auth", "success", &authDBStatus)
	if err != nil {
		t.Fatalf("health.NewHealth(auth) error = %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/_health", nil)
	Health(healthServiceStub{
		gateway:    &gateway,
		gatewayErr: errs.ErrUnavailable,
		auth:       &authHealth,
	}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}

	var response dto.HealthResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Services) != 2 {
		t.Fatalf("services count = %d, want 2", len(response.Services))
	}
	if response.Services[0].Status != "error" {
		t.Fatalf("gateway status = %q, want error", response.Services[0].Status)
	}
}

func TestHealthReturnsOKWhenDependenciesAreHealthy(t *testing.T) {
	ctx := context.Background()
	status := "success"
	gateway, _ := health.NewHealth(ctx, "gateway", "success", &status)
	authHealth, _ := health.NewHealth(ctx, "auth", "success", &status)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/_health", nil)
	Health(healthServiceStub{gateway: &gateway, auth: &authHealth}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestHealthReturnsServiceUnavailableForAuthFailure(t *testing.T) {
	ctx := context.Background()
	status := "success"
	errorStatus := "error"
	gateway, _ := health.NewHealth(ctx, "gateway", "success", &status)
	authHealth, _ := health.NewHealth(ctx, "auth", "error", &errorStatus)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/_health", nil)
	Health(healthServiceStub{
		gateway: &gateway,
		auth:    &authHealth,
		authErr: errors.New("auth unavailable"),
	}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}
