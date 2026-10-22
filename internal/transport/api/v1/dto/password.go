package dto

import (
	"context"
	"strings"

	router_err "github.com/Weit145/gateway-project-service/internal/transport/api/v1/err"
	"github.com/google/uuid"
)

type PasswordRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=32"`
}

func (r *PasswordRequest) Validate(ctx context.Context) error {
	r.Email = strings.TrimSpace(r.Email)

	if err := router_err.Struct(ctx, r); err != nil {
		return err
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

	if err := router_err.Struct(ctx, r); err != nil {
		return err
	}

	return nil
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=6,max=32"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=32"`
}

func (r *ChangePasswordRequest) Validate(ctx context.Context) error {
	if err := router_err.Struct(ctx, r); err != nil {
		return err
	}

	return nil
}
