package dto

import (
	"context"
	"strings"

	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,min=4,max=32"`
	Password string `json:"password" validate:"required,min=6,max=32"`
}

func (r *LoginRequest) Validate(ctx context.Context) error {
	r.Email = strings.TrimSpace(r.Email)
	if err := router_err.Struct(ctx, r); err != nil {
		return err
	}

	return nil
}
