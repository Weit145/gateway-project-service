package dto

import (
	"context"
	"strings"

	"github.com/Weit145/gateway-project-service/internal/domain/auth"
	router_err "github.com/Weit145/gateway-project-service/internal/transport/api/v1/err"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=32"`
}

func (r *RegisterRequest) Validate(ctx context.Context) error {
	r.Email = strings.TrimSpace(r.Email)

	if err := router_err.Struct(ctx, r); err != nil {
		return err
	}

	return nil
}

type ConfirmEmailRequest struct {
	Token  string    `json:"token" validate:"required"`
	UserId uuid.UUID `json:"userId" validate:"required"`
}

type ConfirmEmailResponse struct {
	AccessToken string `json:"access_token"`
}

func NewConfirmEmailResponse(accessToken auth.Token) *ConfirmEmailResponse {
	return &ConfirmEmailResponse{
		AccessToken: accessToken.Token(),
	}
}

func (r *ConfirmEmailRequest) Validate(ctx context.Context) error {
	r.Token = strings.TrimSpace(r.Token)

	if err := router_err.Struct(ctx, r); err != nil {
		return err
	}

	return nil
}

type RegisterCompleteRequest struct {
	Password string `json:"password" validate:"required,min=6,max=32"`
}

type RegisterCompleteResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

func NewRegisterResponse(user auth.User, token auth.Token) RegisterCompleteResponse {
	return RegisterCompleteResponse{
		AccessToken: token.Token(),
		User:        DomainToUserDTO(user),
	}
}

func (r *RegisterCompleteRequest) Validate(ctx context.Context) error {
	if err := router_err.Struct(ctx, r); err != nil {
		return err
	}

	return nil
}

type RepeatEmailRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=32"`
}

func (r *RepeatEmailRequest) Validate(ctx context.Context) error {
	r.Email = strings.TrimSpace(r.Email)
	if err := router_err.Struct(ctx, r); err != nil {
		return err
	}

	return nil
}
