package handler_auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/middleware"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/request"
	"github.com/google/uuid"
)

type LoginService interface {
	Login(ctx context.Context, email string, password string) (auth.User, auth.Tokens, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
}

// Login godoc
// @Summary Login user
// @Description Authenticates a user and returns an access token. Refresh token is set in an HttpOnly cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 201 {object} dto.RegisterCompleteResponse
// @Header 201 {string} Set-Cookie "refresh_token cookie"
// @Failure 400 {object} router_err.Response
// @Failure 401 {object} router_err.Response
// @Failure 403 {object} router_err.Response
// @Failure 500 {object} router_err.Response
// @Router /auth/login [post]
func Login(ser LoginService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		var req dto.LoginRequest
		if err := request.DecodeJSON(w, r, &req); err != nil {
			slog.DebugContext(ctx, "failed to decode request body", slog.Any("error", err))
			router_err.DecodeFailed(w, err)
			return
		}

		if err := req.Validate(ctx); err != nil {
			slog.DebugContext(ctx, "failed to validate request body", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}
		ctx = logger.WithLogEmail(ctx, req.Email)

		user, tokens, err := ser.Login(ctx, req.Email, req.Password)
		if err != nil {
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed to login user", slog.Any("error", err))
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
		request.WriteJSON(w, http.StatusCreated, resp)
	}
}

// Logout godoc
// @Summary Logout user
// @Description Requires "Bearer <access_token>" in Authorization and revokes the refresh_token from the HttpOnly cookie.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} router_err.Response
// @Header 200 {string} Set-Cookie "Clears the refresh_token cookie"
// @Failure 400 {object} router_err.Response
// @Failure 401 {object} router_err.Response
// @Failure 500 {object} router_err.Response
// @Router /auth/logout [post]
func Logout(ser LoginService) middleware.AuthHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {

		ctx := r.Context()

		cookie, err := r.Cookie("refresh_token")
		if err != nil || cookie.Value == "" {
			slog.DebugContext(ctx, "cookie not found", slog.Any("error", err))
			router_err.CookieFaild(w)
			return
		}

		refreshToken := cookie.Value
		if err := ser.Logout(ctx, userID, refreshToken); err != nil {
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed to Logout user", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}

		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "refresh_token",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   -1,
			},
		)
		request.WriteJSON(w, http.StatusOK, router_err.Response{Status: "ok"})

		slog.DebugContext(ctx, "successful Logout")
	}
}
