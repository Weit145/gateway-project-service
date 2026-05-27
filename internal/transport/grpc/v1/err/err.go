package mapper_err_grpc

import (
	"context"
	"fmt"
	"log/slog"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Field struct {
	Name  string
	Value string
}

func MapErr(err error) error {
	switch status.Code(err) {
	case codes.Unauthenticated:
		return errs.ErrUnauthorized
	case codes.PermissionDenied:
		return errs.ErrForbidden
	case codes.InvalidArgument:
		return errs.ErrInvalidInput
	case codes.NotFound:
		return errs.ErrNotFound
	case codes.AlreadyExists:
		return errs.ErrAlreadyExists
	default:
		return errs.ErrInternal
	}
}

func InvalidErr(ctx context.Context, fields ...Field) error {
	for _, field := range fields {
		if field.Value == "" {
			slog.DebugContext(ctx, "Received invalid "+field.Name+" from gRPC response")
			return fmt.Errorf(field.Name+" is required: %w", errs.ErrInvalidInput)
		}
	}
	return nil
}
