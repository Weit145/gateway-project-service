package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/domain/health"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
)

type HealthService interface {
	CheckHealthGateway(ctx context.Context) (*health.Health, error)
	CheckHealthAuth(ctx context.Context) (*health.Health, error)
}

// Health godoc
// @Summary Check health
// @Description Checks gateway and auth service availability.
// @Tags health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 503 {object} dto.HealthResponse
// @Router /_health [get]
func Health(ser HealthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		gatewayHealth, gatewayErr := ser.CheckHealthGateway(ctx)
		if gatewayErr != nil {
			logCtx := logger.ErrorCtx(ctx, gatewayErr)
			slog.WarnContext(logCtx, "gateway is not ready", slog.Any("error", gatewayErr))
		}

		authHealth, authErr := ser.CheckHealthAuth(ctx)
		if authErr != nil {
			logCtx := logger.ErrorCtx(ctx, authErr)
			slog.WarnContext(logCtx, "auth is not ready", slog.Any("error", authErr))
		}

		response := dto.NewHealthResponse(gatewayHealth, authHealth)
		status := http.StatusOK
		if gatewayErr != nil || authErr != nil {
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(response)
	}
}
