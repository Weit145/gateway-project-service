package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
	"github.com/Weit145/gateway-startup/internal/usecase"
)

// Health godoc
// @Summary Check health
// @Description Checks gateway and auth service availability.
// @Tags health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 500 {object} dto.Response
// @Router /_health [get]
func Health(ser usecase.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		gatewayHealth, err := ser.CheckHealthGateway(ctx)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to check health gateway", "error", err)
			router_err.MapErr(w, r, err)
			return
		}

		authHealth, err := ser.CheckHealthAuth(ctx)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to check health", "error", err)

			router_err.MapErr(w, r, err)
			return
		}
		slog.DebugContext(ctx, "successful _health")

		response := dto.NewHealthResponse(gatewayHealth, authHealth)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
