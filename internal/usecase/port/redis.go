package port

import (
	"context"
)

type RedisServices interface {
	CheckHealth(ctx context.Context) error
}
