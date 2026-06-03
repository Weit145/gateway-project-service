package middleware

import (
	"crypto/rsa"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/config"
	"github.com/Weit145/gateway-startup/internal/jwt"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
	mapper_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "panic", rec)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(dto.Error("internal server error"))
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

func CheckLogin(next http.Handler) http.Handler {
	err, cfg := config.InitConfig()
	if err != nil {
		slog.Error("failed to load JWT public key", "error", err)
		return authConfigErrorMiddleware(next)
	}

	return CheckLoginWithPublicKey(cfg.JWT.PublicKey)(next)
}

func CheckLoginWithPublicKey(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	if publicKey == nil {
		slog.Error("JWT public key is not configured")
		return authConfigErrorMiddleware
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID, err := jwt.CheckToken(r.Header.Get("Authorization"), publicKey)
			if err != nil {
				slog.DebugContext(ctx, "failed to check bearer token", "error", err)
				mapper_err.MapErr(w, r, err)
				return
			}

			ctx = logger.WithLogId(ctx, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authConfigErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mapper_err.MapErr(w, r, errs.ErrInternal)
	})
}
