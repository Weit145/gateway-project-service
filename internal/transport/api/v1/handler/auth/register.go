package handler_auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-project-service/internal/domain/auth"
	"github.com/Weit145/gateway-project-service/internal/logger"
	"github.com/Weit145/gateway-project-service/internal/transport/api/v1/dto"
	router_err "github.com/Weit145/gateway-project-service/internal/transport/api/v1/err"
	"github.com/Weit145/gateway-project-service/internal/transport/api/v1/middleware"
	"github.com/Weit145/gateway-project-service/internal/transport/api/v1/request"
	"github.com/google/uuid"
)

type RegisterService interface {
	Register(ctx context.Context, email string) error
	ConfirmEmail(ctx context.Context, token string, userID uuid.UUID) (auth.Token, error)
	RegisterComplete(ctx context.Context, password string, session uuid.UUID) (auth.User, auth.Tokens, error)
	RepeatEmail(ctx context.Context, email string) error
}

type RateLimiter interface {
	LimiterEmail(ctx context.Context, email string) error
}

// RegisterEmailRequest godoc
// @Summary Request registration email
// @Description Creates a pending account and sends an email verification link.
// @Tags registration
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration email"
// @Success 200 {object} router_err.Response
// @Failure 400 {object} router_err.Response
// @Failure 409 {object} router_err.Response
// @Failure 429 {object} router_err.Response
// @Failure 500 {object} router_err.Response
// @Router /register/request [post]
func RegisterEmailRequest(ser RegisterService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		var req dto.RegisterRequest

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

		if err := ser.Register(ctx, req.Email); err != nil {
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed to register user", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}

		request.WriteJSON(w, http.StatusOK, router_err.Response{Status: "ok"})

	}
}

// RegisterEmailConfirm godoc
// @Summary Confirm registration email
// @Description Validates the email token and returns a short-lived registration access token.
// @Tags registration
// @Accept json
// @Produce json
// @Param request body dto.ConfirmEmailRequest true "Email confirmation"
// @Success 200 {object} dto.ConfirmEmailResponse
// @Failure 400 {object} router_err.Response
// @Failure 401 {object} router_err.Response
// @Failure 403 {object} router_err.Response
// @Failure 500 {object} router_err.Response
// @Router /register/confirm [post]
func RegisterEmailConfirm(ser RegisterService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.ConfirmEmailRequest
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

		ctx = logger.WithLogId(ctx, req.UserId)

		token, err := ser.ConfirmEmail(ctx, req.Token, req.UserId)
		if err != nil {
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed to confirm email", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}
		request.WriteJSON(w, http.StatusOK, dto.NewConfirmEmailResponse(token))
	}
}

// RegisterComplete godoc
// @Summary Complete registration
// @Description Sets the password using the short-lived registration Bearer token and creates an authenticated session.
// @Tags registration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RegisterCompleteRequest true "Password"
// @Success 201 {object} dto.RegisterCompleteResponse
// @Header 201 {string} Set-Cookie "refresh_token cookie"
// @Failure 400 {object} router_err.Response
// @Failure 401 {object} router_err.Response
// @Failure 403 {object} router_err.Response
// @Failure 500 {object} router_err.Response
// @Router /register/complete [post]
func RegisterComplete(ser RegisterService) middleware.SessionHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, session uuid.UUID, userID uuid.UUID) {
		ctx := r.Context()

		var req dto.RegisterCompleteRequest
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

		user, tokens, err := ser.RegisterComplete(ctx, req.Password, session)
		if err != nil {
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed to complete registration", slog.Any("error", err))
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

// RepeatEmail godoc
// @Summary Repeat registration email
// @Description Invalidates the previous verification token and sends a new verification link.
// @Tags registration
// @Accept json
// @Produce json
// @Param request body dto.RepeatEmailRequest true "Registration email"
// @Success 200 {object} router_err.Response
// @Failure 400 {object} router_err.Response
// @Failure 404 {object} router_err.Response
// @Failure 429 {object} router_err.Response
// @Failure 500 {object} router_err.Response
// @Router /register/repeat-email [post]
func RepeatEmail(ser RegisterService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.RepeatEmailRequest

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

		if err := ser.RepeatEmail(ctx, req.Email); err != nil {
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed to repeat email", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}

		request.WriteJSON(w, http.StatusOK, router_err.Response{Status: "ok"})
	}
}
