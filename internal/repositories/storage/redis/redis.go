package redis

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/redis/go-redis/v9"
)

type Setting struct {
	Addr     string
	Password string
	DB       int
	Protocol int
}

type Repository struct {
	db *redis.Client
}

func NewRepository(ctx context.Context, setting Setting) (*Repository, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     setting.Addr,
		Password: setting.Password,
		DB:       setting.DB,
		Protocol: setting.Protocol,
	})

	err := rdb.Set(ctx, "foo", "bar", 0).Err()
	if err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("set redis: %w", err))
	}

	_, err = rdb.Get(ctx, "foo").Result()
	if err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("get redis: %w", err))
	}
	return &Repository{
		db: rdb,
	}, nil
}
