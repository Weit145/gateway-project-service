package port

import (
	"context"

	"github.com/google/uuid"
)

type RedisServices interface {
	CheckHealth(ctx context.Context) error
	LimiterIP(ctx context.Context, ip string) error
	LimiterUserID(ctx context.Context, userID uuid.UUID) error
	LimiterEmail(ctx context.Context, email string) error
}
