package handler_auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
	"github.com/Weit145/gateway-startup/internal/usecase"
)

// RefreshToken godoc
// @Summary Refresh access token
// @Description Validates the refresh_token HttpOnly cookie and returns a new access token.
// @Tags auth
// @Produce json
// @Param Cookie header string true "refresh_token=<token>"
// @Success 200 {object} dto.RegisterResponse
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/refresh-token [post]
func RefreshToken(ser usecase.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cookie, err := r.Cookie("refresh_token")
		if err != nil || cookie.Value == "" {
			slog.DebugContext(ctx, "cookie not found")
			router_err.CookieFaild(w)
			return
		}

		refreshToken := cookie.Value

		user, tokens, err := ser.RefreshToken(ctx, refreshToken)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to refresh token", "error", err)
			router_err.MapErr(w, r, err)
			return
		}

		slog.DebugContext(ctx, "successful token refresh")
		resp := dto.NewRegisterResponse(user, tokens)

		w.Header().Set("Content-Type", "application/json")
		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "refresh_token",
				Value:    tokens.RefreshToken(),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			},
		)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
