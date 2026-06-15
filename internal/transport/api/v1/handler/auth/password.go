package handler_auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/authctx"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/request"
	"github.com/google/uuid"
)

type PasswordService interface {
	PasswordRequest(ctx context.Context, email string) error
	PasswordConfirm(ctx context.Context, token string, userID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error
}

func PasswordRequest(ser PasswordService, rd RateLimiter) http.HandlerFunc {
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

		if err := rd.LimiterEmail(ctx, req.Email); err != nil {
			slog.DebugContext(ctx, "rate limit email", slog.Any("error", err))
			router_err.MapErr(w, r, err)
			return
		}

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

		if err := ser.PasswordConfirm(ctx, req.Token, req.UserId); err != nil {
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

func ChangePassword(ser PasswordService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.ChagePasswordRequest

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

		userID, ok := authctx.UserID(ctx)
		if !ok {
			slog.WarnContext(ctx, "user ID not found in context")
			router_err.MapErr(w, r, errs.ErrUnauthorized)
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
