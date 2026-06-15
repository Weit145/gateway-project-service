package my_redis

import (
	"context"
	"errors"
	"net"
	"testing"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/go-redis/redis_rate/v10"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestLimiterRedisFailureIsUnavailable(t *testing.T) {
	dialErr := errors.New("redis dial failed")
	client := redis.NewClient(&redis.Options{
		Addr:       "redis.invalid:6379",
		MaxRetries: -1,
		Dialer: func(context.Context, string, string) (net.Conn, error) {
			return nil, dialErr
		},
	})
	t.Cleanup(func() {
		_ = client.Close()
	})

	repository := &Repository{
		db:      client,
		limiter: redis_rate.NewLimiter(client),
	}

	tests := []struct {
		name string
		call func() error
	}{
		{name: "IP", call: func() error { return repository.LimiterIP(context.Background(), "127.0.0.1") }},
		{name: "user", call: func() error { return repository.LimiterUserID(context.Background(), uuid.New()) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if !errors.Is(err, errs.ErrUnavailable) {
				t.Fatalf("limiter error = %v, want ErrUnavailable", err)
			}
		})
	}
}
