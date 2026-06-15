package my_redis

import (
	"context"
	"fmt"

	"github.com/Weit145/gateway-startup/internal/logger"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type Setting struct {
	Addr     string
	Password string
	DB       int
	Protocol int
}

type Repository struct {
	db      *redis.Client
	limiter *redis_rate.Limiter
}

func NewRepository(ctx context.Context, setting Setting) (*Repository, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     setting.Addr,
		Password: setting.Password,
		DB:       setting.DB,
		Protocol: setting.Protocol,
	})

	limiter := redis_rate.NewLimiter(rdb)
	_, err := limiter.Allow(ctx, "project:123", redis_rate.PerSecond(10))
	if err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("rate_limit redis: %w", err))
	}

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("ping redis: %w", err))
	}

	return &Repository{
		db:      rdb,
		limiter: limiter,
	}, nil
}

func (r *Repository) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

func (r *Repository) CheckHealth(ctx context.Context) error {
	if err := r.db.Ping(ctx).Err(); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("ping redis: %w", err))
	}
	return nil
}
