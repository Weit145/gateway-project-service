package mapper_err_grpc

import (
	"context"
	"errors"
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

func MapErr(err error) (interface{}, error) {
	switch {
	case errors.Is(err, errs.ErrUnauthorized):
		return nil, status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, errs.ErrForbidden):
		return nil, status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, errs.ErrInvalidInput):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, errs.ErrNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	default:
		return nil, status.Error(codes.Internal, err.Error())
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
