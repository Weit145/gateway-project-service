package port

import (
	"context"

	"github.com/Weit145/gateway-startup/internal/domain/auth"
	"github.com/google/uuid"
)

type GRPServices interface {
	CheckHealth(ctx context.Context) (bool, error)
	Register(ctx context.Context, email string) error
	ConfirmEmail(ctx context.Context, token string, userID string) (auth.Token, error)
	RegisterComplete(ctx context.Context, password string, session uuid.UUID) (auth.User, auth.Tokens, error)
	RepeatEmail(ctx context.Context, email string) error
	Login(ctx context.Context, email string, password string) (auth.User, auth.Tokens, error)
	Logout(ctx context.Context, userID string, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (auth.User, auth.Tokens, error)
	PasswordRequest(ctx context.Context, email string) error
	PasswordConfirm(ctx context.Context, token string, userID uuid.UUID, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error
}
