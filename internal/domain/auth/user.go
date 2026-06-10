package auth

import (
	"context"
	"fmt"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/lib/valid"
	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/google/uuid"
)

type User struct {
	id       uuid.UUID
	email    string
	username string
	role     string
}

func NewUser(ctx context.Context, id uuid.UUID, email string, username string, role string) (User, error) {
	if id == uuid.Nil {
		return User{}, logger.WrapError(ctx, fmt.Errorf("id cannot be empty"))
	}

	email, err := valid.NormalizeEmail(ctx, email)
	if err != nil {
		return User{}, logger.WrapError(ctx, fmt.Errorf("invalid email: %w", err))
	}

	username, err = valid.NormalizeValue(ctx, username)
	if err != nil {
		return User{}, logger.WrapError(ctx, fmt.Errorf("invalid name: %w", err))
	}

	role, err = valid.NormalizeValue(ctx, role)
	if err != nil {
		return User{}, logger.WrapError(ctx, fmt.Errorf("invalid role: %w", err))
	}

	if role != "user" && role != "admin" {
		return User{}, logger.WrapError(ctx, fmt.Errorf("invalid role %q: %w", role, errs.ErrInvalidInput))
	}

	return User{
		id:       id,
		email:    email,
		username: username,
		role:     role,
	}, nil
}

func (u User) ID() uuid.UUID {
	return u.id
}

func (u User) Email() string {
	return u.email
}

func (u User) Username() string {
	return u.username
}

func (u User) Role() string {
	return u.role
}
