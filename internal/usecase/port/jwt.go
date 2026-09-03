package port

import (
	"context"

	"github.com/google/uuid"
)

type JWTServices interface {
	CheckToken(ctx context.Context, authHeader string) (uuid.UUID, uuid.UUID, error)
}
