package port

import (
	"github.com/google/uuid"
)

type JWTServices interface {
	CheckToken(authHeader string) (uuid.UUID, uuid.UUID, error)
}
