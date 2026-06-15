package middleware

import (
	"crypto/rsa"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/jwt"
	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/authctx"
	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
)

func CheckLogin(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	if publicKey == nil {
		slog.Error("JWT public key is not configured")
		return authConfigErrorMiddleware
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID, jwtId, err := jwt.CheckToken(r.Header.Get("Authorization"), publicKey)
			if err != nil {
				slog.DebugContext(ctx, "failed to check bearer token", "error", err)
				router_err.MapErr(w, r, err)
				return
			}

			ctx = authctx.WithUserID(ctx, userID)
			ctx = logger.WithLogId(ctx, userID)
			ctx = logger.WithLogJwtId(ctx, jwtId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authConfigErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router_err.MapErr(w, r, errs.ErrInternal)
	})
}
