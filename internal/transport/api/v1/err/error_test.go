package router_err

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	errs "github.com/Weit145/gateway-project-service/internal/lib/e"
)

func TestMapErrDependencyFailures(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "unavailable", err: errs.ErrUnavailable, wantStatus: http.StatusServiceUnavailable},
		{name: "timeout", err: errs.ErrTimeout, wantStatus: http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			MapErr(recorder, request, tt.err)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}

func TestDecodeFailed(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "oversized body",
			err:        &http.MaxBytesError{Limit: 10},
			wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:       "invalid JSON",
			err:        errors.New("invalid JSON"),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			DecodeFailed(recorder, tt.err)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
