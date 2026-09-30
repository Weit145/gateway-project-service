package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

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
				slog.Error("panic recovered",
					"panic", rec,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(router_err.Error("internal server error"))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		slog.Debug(
			"HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start),
		)
	})
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		id, err := uuid.Parse(r.Header.Get("X-Request-ID"))
		if err != nil {
			id = uuid.New()
		}

		ctx := logger.WithLogRequestID(r.Context(), id)

		w.Header().Set("X-Request-ID", id.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
