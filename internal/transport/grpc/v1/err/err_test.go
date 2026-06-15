package mapper_err_grpc

import (
	"errors"
	"testing"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapErr(t *testing.T) {
	tests := []struct {
		name string
		code codes.Code
		want error
	}{
		{name: "deadline exceeded", code: codes.DeadlineExceeded, want: errs.ErrTimeout},
		{name: "unavailable", code: codes.Unavailable, want: errs.ErrUnavailable},
		{name: "invalid argument", code: codes.InvalidArgument, want: errs.ErrInvalidInput},
		{name: "internal", code: codes.Internal, want: errs.ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapErr(status.Error(tt.code, tt.name))
			if !errors.Is(got, tt.want) {
				t.Fatalf("MapErr() = %v, want %v", got, tt.want)
			}
		})
	}
}
