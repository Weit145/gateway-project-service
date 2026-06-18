package handler_auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/middleware"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/request"
	"github.com/google/uuid"
)

type PasswordService interface {
	PasswordRequest(ctx context.Context, email string) error
	PasswordConfirm(ctx context.Context, token string, userID uuid.UUID, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error
}

// PasswordRequest godoc
// @Summary Request password reset
// @Description Sends a password reset request to the user's email.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.PasswordRequest true "Password reset request"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 429 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/password-reset/request [post]
func PasswordRequest(ser PasswordService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.PasswordRequest

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

		if err := ser.PasswordRequest(ctx, req.Email); err != nil {
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed password request", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(dto.Response{Status: "ok"})

		slog.DebugContext(ctx, "successful password request")
	}
}

// PasswordConfirm godoc
// @Summary Confirm password reset
// @Description Confirms a password reset by token and sets a new password.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.PasswordConfirmRequest true "Password reset confirmation"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/password-reset/confirm [post]
func PasswordConfirm(ser PasswordService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.PasswordConfirmRequest

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

		if err := ser.PasswordConfirm(ctx, req.Token, req.UserId, req.Password); err != nil {
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed password request", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(dto.Response{Status: "ok"})

		slog.DebugContext(ctx, "successful password request")
	}
}

// ChangePassword godoc
// @Summary Change password
// @Description Changes the current user's password. Requires "Bearer <access_token>" in Authorization.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "Change password request"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Failure 429 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /auth/change-password [post]
func ChangePassword(ser PasswordService) middleware.AuthHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
		ctx := r.Context()

		var req dto.ChangePasswordRequest

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

		ctx = logger.WithLogId(ctx, userID)

		if err := ser.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword); err != nil {
			logCtx := logger.ErrorCtx(ctx, err)
			slog.ErrorContext(logCtx, "failed password request", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(dto.Response{Status: "ok"})

		slog.DebugContext(ctx, "successful password request")
	}
}
