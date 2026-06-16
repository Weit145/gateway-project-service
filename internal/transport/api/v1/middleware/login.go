package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/logger"
	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
	"github.com/google/uuid"
)

type LoginService interface {
	CheckLogin(ctx context.Context, authHeader string) (uuid.UUID, uuid.UUID, error)
}

func CheckLogin(ser LoginService) func(AuthHandlerFunc) http.Handler {
	return func(next AuthHandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID, jwtId, err := ser.CheckLogin(ctx, r.Header.Get("Authorization"))
			if err != nil {
				slog.DebugContext(ctx, "failed to check bearer token", slog.Any("error", err))
				router_err.MapErr(w, r, err)
				return
			}

			ctx = logger.WithLogId(ctx, userID)
			ctx = logger.WithLogJwtId(ctx, jwtId)
			next(w, r.WithContext(ctx), userID)
		})
	}
}
