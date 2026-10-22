package my_redis

import (
	"context"
	"fmt"
	"strings"

	errs "github.com/Weit145/gateway-project-service/internal/lib/e"
	"github.com/Weit145/gateway-project-service/internal/logger"
	"github.com/go-redis/redis_rate/v10"
	"github.com/google/uuid"
)

func (r *Repository) LimiterIP(ctx context.Context, ip string) error {
	key := "rate_limit:ip:" + ip
	res, err := r.limiter.Allow(ctx, key, redis_rate.PerSecond(3))
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("allow IP rate limit: %w: %w", err, errs.ErrUnavailable))
	}
	if res.Allowed == 0 {
		return errs.ErrRateLimit
	}
	return nil
}

func (r *Repository) LimiterUserID(ctx context.Context, userID uuid.UUID) error {
	key := "rate_limit:user:id:" + userID.String()
	res, err := r.limiter.Allow(ctx, key, redis_rate.PerSecond(3))
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("allow user rate limit: %w: %w", err, errs.ErrUnavailable))
	}
	if res.Allowed == 0 {
		return errs.ErrRateLimit
	}
	return nil
}
func (r *Repository) LimiterEmail(ctx context.Context, email string) error {

	email = strings.TrimSpace(email)
	if len(email) > 5 {
		email = email[:5] + strings.Repeat("*", len(email)-5)
	}
	key := "rate_limit:email:" + email
	res, err := r.limiter.Allow(ctx, key, redis_rate.PerMinute(2))
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("allow user rate limit: %w: %w", err, errs.ErrUnavailable))
	}
	if res.Allowed == 0 {
		return errs.ErrRateLimit
	}
	return nil
}
