package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/authctx"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
	mapper_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
	"github.com/Weit145/gateway-startup/internal/usecase"
)

// Register godoc
// @Summary Register user
// @Description Registers a new user and sends an email verification request.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 409 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/register [post]
func Register(ser usecase.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		var req dto.RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.DebugContext(ctx, "failed to decode request body", "error", err)

			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.Response{Status: "invalid request"})
			return
		}

		if err := req.Validate(ctx); err != nil {
			mapper_err.MapErr(w, r, err)
			return
		}

		ctx = logger.WithLogEmail(ctx, req.Email)

		err = ser.Register(ctx, req.Email, req.Password, req.Name)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to register user", "error", err)

			mapper_err.MapErr(w, r, err)
			return
		}

		slog.DebugContext(ctx, "successful register")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(dto.Response{Status: "ok"})
	}
}

// Login godoc
// @Summary Login user
// @Description Authenticates a user and returns an access token. Refresh token is set in an HttpOnly cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 201 {object} dto.RegisterResponse
// @Header 201 {string} Set-Cookie "refresh_token cookie"
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/login [post]
func Login(ser usecase.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		var req dto.LoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.DebugContext(ctx, "failed to decode request body", "error", err)

			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.Response{Status: "invalid request"})
			return
		}

		if err := req.Validate(ctx); err != nil {
			mapper_err.MapErr(w, r, err)
			return
		}

		ctx = logger.WithLogEmail(ctx, req.Email)

		user, tokens, err := ser.Login(ctx, req.Email, req.Password)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to login user", "error", err)

			mapper_err.MapErr(w, r, err)
			return
		}

		slog.DebugContext(ctx, "successful login")
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

// ConfirmEmail godoc
// @Summary Confirm user email
// @Description Confirms a user email by verification token and returns an access token. Refresh token is set in an HttpOnly cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ConfirmEmailRequest true "Confirm email request"
// @Success 201 {object} dto.RegisterResponse
// @Header 201 {string} Set-Cookie "refresh_token cookie"
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/confirm-email [post]
func ConfirmEmail(ser usecase.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.ConfirmEmailRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.DebugContext(ctx, "failed to decode request body", "error", err)

			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.Response{Status: "invalid request"})
			return
		}

		if err := req.Validate(ctx); err != nil {
			mapper_err.MapErr(w, r, err)
			return
		}

		ctx = logger.WithLogId(ctx, req.UserId)

		user, tokens, err := ser.ConfirmEmail(ctx, req.Token, req.UserId)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to confirm email", "error", err)

			mapper_err.MapErr(w, r, err)
			return
		}

		slog.DebugContext(ctx, "successful email confirmation")
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

// Logout godoc
// @Summary Logout user
// @Description Requires "Bearer <access_token>" in Authorization and revokes the refresh_token from the HttpOnly cookie.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Header 200 {string} Set-Cookie "Clears the refresh_token cookie"
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/logout [post]
func Logout(ser usecase.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		userID, ok := authctx.UserID(ctx)
		if !ok {
			slog.ErrorContext(ctx, "user ID not found in context")
			mapper_err.MapErr(w, r, errs.ErrUnauthorized)
			return
		}

		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			slog.DebugContext(ctx, "cookie not found")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.Response{Status: "refresh token is required"})
			return
		}

		refreshToken := cookie.Value
		if refreshToken == "" {
			slog.DebugContext(ctx, "refresh token not found in cookie")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.Response{Status: "refresh token is required"})
			return
		}

		err = ser.Logout(ctx, userID, refreshToken)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to Logout user", "error", err)
			mapper_err.MapErr(w, r, err)
			return
		}

		slog.DebugContext(ctx, "successful Logout")

		w.Header().Set("Content-Type", "application/json")
		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "refresh_token",
				Value:    "",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   -1,
			},
		)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(dto.Response{Status: "ok"})
	}
}

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
		if err != nil {
			slog.DebugContext(ctx, "cookie not found")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.Response{Status: "refresh token is required"})
			return
		}

		refreshToken := cookie.Value
		if refreshToken == "" {
			slog.DebugContext(ctx, "refresh token not found in cookie")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.Response{Status: "refresh token is required"})
			return
		}

		user, tokens, err := ser.RefreshToken(ctx, refreshToken)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to refresh token", "error", err)
			mapper_err.MapErr(w, r, err)
			return
		}

		slog.DebugContext(ctx, "successful token refresh")
		resp := dto.NewRegisterResponse(user, tokens)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
