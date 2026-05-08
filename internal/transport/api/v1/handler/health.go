package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
	mapererr_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
	"github.com/Weit145/gateway-startup/internal/usecase"
)

// Health godoc
// @Summary Check health
// @Description Checks gateway and auth service availability.
// @Tags health
// @Produce json
// @Success 200 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /_health [get]
func Health(ser usecase.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ser.CheckHealth(r.Context())
		if err != nil {
			slog.Error("failed to check health", "error", err)

			mapererr_err.MapErr(w, r, err)
			return
		}
		slog.Debug("successful _health")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(dto.Response{Status: "ok"})
	}
}
