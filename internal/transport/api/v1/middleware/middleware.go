package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/logger"
	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
	"github.com/google/uuid"
)

type RateLimiter interface {
	LimiterIP(ctx context.Context, ip string) error
	LimiterUserID(ctx context.Context, userID uuid.UUID) error
}

type AuthHandlerFunc func(http.ResponseWriter, *http.Request, uuid.UUID)
type SessionHandlerFunc func(http.ResponseWriter, *http.Request, uuid.UUID, uuid.UUID)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "panic", rec)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(router_err.Error("internal server error"))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		ctx := logger.WithLogIp(r.Context(), ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
