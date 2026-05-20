package dto

import (
	"context"
	"log/slog"
	"strings"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/google/uuid"
)

type User struct {
	Id    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
	Role  string    `json:"role"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=4,max=32"`
	Email    string `json:"email" validate:"required,email,min=4,max=32"`
	Password string `json:"password" validate:"required,min=6,max=32"`
}

type RegisterResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

func (r *RegisterRequest) Validate(ctx context.Context) error {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)

	if len(r.Name) < 4 || len(r.Name) > 32 {
		slog.DebugContext(ctx, "Validation failed Name")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	if len(r.Email) < 4 || len(r.Email) > 32 || !strings.Contains(r.Email, "@") {
		slog.DebugContext(ctx, "Validation failed Email")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	if len(r.Password) < 6 || len(r.Password) > 32 {
		slog.DebugContext(ctx, "Validation failed Password")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	return nil
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,min=4,max=32"`
	Password string `json:"password" validate:"required,min=6,max=32"`
}

func (r *LoginRequest) Validate(ctx context.Context) error {
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)

	if len(r.Email) < 4 || len(r.Email) > 32 || !strings.Contains(r.Email, "@") {
		slog.DebugContext(ctx, "Validation failed Email")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	if len(r.Password) < 6 || len(r.Password) > 32 {
		slog.DebugContext(ctx, "Validation failed Password")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	return nil
}
