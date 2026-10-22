package handler_auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-project-service/internal/domain/auth"
	"github.com/Weit145/gateway-project-service/internal/logger"
	"github.com/Weit145/gateway-project-service/internal/transport/api/v1/dto"
	router_err "github.com/Weit145/gateway-project-service/internal/transport/api/v1/err"
	"github.com/Weit145/gateway-project-service/internal/transport/api/v1/request"
)

type RefreshService interface {
	RefreshToken(ctx context.Context, refreshToken string) (auth.User, auth.Tokens, error)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Validates the refresh_token HttpOnly cookie and returns a new access token.
// @Tags auth
// @Produce json
// @Param Cookie header string true "refresh_token=<token>"
// @Success 200 {object} dto.RegisterCompleteResponse
// @Failure 400 {object} router_err.Response
// @Failure 401 {object} router_err.Response
// @Failure 404 {object} router_err.Response
// @Failure 500 {object} router_err.Response
// @Router /auth/refresh-token [post]
func RefreshToken(ser RefreshService) http.HandlerFunc {
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
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed to refresh token", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}

		resp := dto.NewRegisterResponse(user, tokens.AccessToken())

		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "refresh_token",
				Value:    tokens.RefreshToken().Token(),
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			},
		)
		request.WriteJSON(w, http.StatusOK, resp)
	}
}
