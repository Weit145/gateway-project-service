package health

import (
	"context"

	errs "github.com/Weit145/gateway-project-service/internal/lib/e"
	"github.com/Weit145/gateway-project-service/internal/logger"
)

type Health struct {
	name   string
	status string
	db     *string
}

func NewHealth(ctx context.Context, name string, status string, db *string) (Health, error) {
	if name == "" {
		return Health{}, logger.WrapError(ctx, errs.ErrInternal)
	}

	if status == "" {
		return Health{}, logger.WrapError(ctx, errs.ErrInternal)
	}

	return Health{
		name:   name,
		status: status,
		db:     db,
	}, nil
}

func NewHealthList() []Health                               { return []Health{} }
func AddHealth(healths []Health, health ...Health) []Health { return append(healths, health...) }

func (h Health) Name() string   { return h.name }
func (h Health) Status() string { return h.status }
func (h Health) DB() *string    { return h.db }
