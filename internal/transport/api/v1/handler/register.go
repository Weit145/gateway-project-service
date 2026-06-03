package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/logger"
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

		err, resp, refreshToken := ser.Login(ctx, req.Email, req.Password)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to login user", "error", err)

			mapper_err.MapErr(w, r, err)
			return
		}

		slog.DebugContext(ctx, "successful login")

		w.Header().Set("Content-Type", "application/json")
		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "refresh_token",
				Value:    refreshToken,
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

		err, resp, refreshToken := ser.ConfirmEmail(ctx, req.Token, req.UserId)
		if err != nil {
			slog.ErrorContext(logger.ErrorCtx(ctx, err), "failed to confirm email", "error", err)

			mapper_err.MapErr(w, r, err)
			return
		}

		slog.DebugContext(ctx, "successful email confirmation")

		w.Header().Set("Content-Type", "application/json")
		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "refresh_token",
				Value:    refreshToken,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			},
		)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
