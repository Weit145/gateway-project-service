package middleware

import (
	"context"
	"log/slog"
	"net/http"

	router_err "github.com/Weit145/gateway-project-service/internal/transport/api/v1/err"
	"github.com/google/uuid"
)

type ServiceMiddlewareRegistration interface {
	CheckCompleteRegistration(ctx context.Context, authHeader string) (uuid.UUID, uuid.UUID, error)
}

func CheckCompleteRegistration(ser ServiceMiddlewareRegistration) func(SessionHandlerFunc) http.Handler {
	return func(next SessionHandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			session, userID, err := ser.CheckCompleteRegistration(ctx, r.Header.Get("Authorization"))
			if err != nil {
				slog.DebugContext(ctx, "failed to check bearer token", slog.Any("error", err))
				router_err.MapErr(w, r, err)
				return
			}

			next(w, r.WithContext(ctx), session, userID)
		})
	}
}
