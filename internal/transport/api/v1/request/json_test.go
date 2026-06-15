package request

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type decodeRequest struct {
	Name string `json:"name"`
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "unknown field",
			body: `{"name":"value","unknown":true}`,
		},
		{
			name: "multiple JSON values",
			body: `{"name":"value"} {"name":"second"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			if err := DecodeJSON(rec, req, &decodeRequest{}); err == nil {
				t.Fatal("DecodeJSON() error = nil")
			}
		})
	}
}

func TestDecodeJSONRejectsOversizedBody(t *testing.T) {
	body := `{"name":"` + strings.Repeat("a", int(MaxJSONBodySize)) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	err := DecodeJSON(rec, req, &decodeRequest{})
	var maxBytesErr *http.MaxBytesError
	if !errors.As(err, &maxBytesErr) {
		t.Fatalf("DecodeJSON() error = %v, want *http.MaxBytesError", err)
	}
}

func TestDecodeJSONAcceptsSingleKnownObject(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"value"}`))
	rec := httptest.NewRecorder()
	var dst decodeRequest

	if err := DecodeJSON(rec, req, &dst); err != nil {
		t.Fatalf("DecodeJSON() error = %v", err)
	}
	if dst.Name != "value" {
		t.Fatalf("decoded name = %q, want %q", dst.Name, "value")
	}
}
