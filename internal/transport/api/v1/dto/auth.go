package dto

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
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
	Token    string    `json:"token" validate:"required"`
	UserId   uuid.UUID `json:"userId" validate:"required"`
	Name     string    `json:"name" validate:"required,min=4,max=32"`
	Password string    `json:"password" validate:"required,min=6,max=32"`
}

func (r *RegisterRequest) Validate(ctx context.Context) error {
	r.Token = strings.TrimSpace(r.Token)
	r.Name = strings.TrimSpace(r.Name)

	if r.Token == "" {
		slog.DebugContext(ctx, "Validation failed Token")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	if r.UserId == uuid.Nil {
		slog.DebugContext(ctx, "Validation failed UserId")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	if len(r.Name) < 4 || len(r.Name) > 32 {
		slog.DebugContext(ctx, "Validation failed Name")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	if len(r.Password) < 6 || len(r.Password) > 32 {
		slog.DebugContext(ctx, "Validation failed Password")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	return nil
}

type RegisterResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

func NewRegisterResponse(user auth.User, tokens auth.Tokens) RegisterResponse {
	return RegisterResponse{
		AccessToken: tokens.AccessToken(),
		User: User{
			Id:    user.ID(),
			Email: user.Email(),
			Name:  user.Username(),
			Role:  user.Role(),
		},
	}
}

type ConfirmEmailRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=32"`
}

func (r *ConfirmEmailRequest) Validate(ctx context.Context) error {

	r.Email = strings.TrimSpace(r.Email)

	if len(r.Email) < 4 || len(r.Email) > 32 || !strings.Contains(r.Email, "@") {
		slog.DebugContext(ctx, "Validation failed Email")
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

type LogoutRequest struct {
}

type PasswordRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=32"`
}

func (r *PasswordRequest) Validate(ctx context.Context) error {
	r.Email = strings.TrimSpace(r.Email)

	if len(r.Email) < 4 || len(r.Email) > 32 || !strings.Contains(r.Email, "@") {
		slog.DebugContext(ctx, "Validation failed Email")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	return nil
}

type PasswordConfirmRequest struct {
	Token    string    `json:"token" validate:"required"`
	UserId   uuid.UUID `json:"userId" validate:"required"`
	Password string    `json:"password" validate:"required,min=6,max=32"`
}

func (r *PasswordConfirmRequest) Validate(ctx context.Context) error {
	r.Token = strings.TrimSpace(r.Token)

	if r.Token == "" {
		slog.DebugContext(ctx, "Validation failed Token")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	if r.UserId == uuid.Nil {
		slog.DebugContext(ctx, "Validation failed UserId")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	if len(r.Password) < 6 || len(r.Password) > 32 {
		slog.DebugContext(ctx, "Validation failed Password")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	return nil
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=6,max=32"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=32"`
}

func (r *ChangePasswordRequest) Validate(ctx context.Context) error {

	if len(r.OldPassword) < 6 || len(r.OldPassword) > 32 {
		slog.DebugContext(ctx, "Validation failed old Password")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	if len(r.NewPassword) < 6 || len(r.NewPassword) > 32 {
		slog.DebugContext(ctx, "Validation failed new Password")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	return nil
}

type RepeatEmailRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=32"`
}

func (r *RepeatEmailRequest) Validate(ctx context.Context) error {
	r.Email = strings.TrimSpace(r.Email)

	if len(r.Email) < 4 || len(r.Email) > 32 || !strings.Contains(r.Email, "@") {
		slog.DebugContext(ctx, "Validation failed Email")
		return logger.WrapError(ctx, errs.ErrInvalidInput)
	}

	return nil
}
