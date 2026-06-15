package router_err

import (
	"encoding/json"
	"errors"
	"net/http"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/dto"
)

func MapErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errs.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, dto.Error("unauthorized"))
	case errors.Is(err, errs.ErrForbidden):
		writeJSON(w, http.StatusForbidden, dto.Error("forbidden"))
	case errors.Is(err, errs.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, dto.Error("invalid request"))
	case errors.Is(err, errs.ErrNotFound):
		writeJSON(w, http.StatusNotFound, dto.Error("not found"))
	case errors.Is(err, errs.ErrAlreadyExists):
		writeJSON(w, http.StatusConflict, dto.Error("already exists"))
	case errors.Is(err, errs.ErrRateLimit):
		writeJSON(w, http.StatusTooManyRequests, dto.Error("rate limit exceeded"))
	case errors.Is(err, errs.ErrUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, dto.Error("service unavailable"))
	case errors.Is(err, errs.ErrTimeout):
		writeJSON(w, http.StatusGatewayTimeout, dto.Error("gateway timeout"))
	default:
		writeJSON(w, http.StatusInternalServerError, dto.Error("internal server error"))
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func DecodeFailed(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		writeJSON(w, http.StatusRequestEntityTooLarge, dto.Error("request body too large"))
		return
	}

	writeJSON(w, http.StatusBadRequest, dto.Error("invalid request"))
}

func CookieFaild(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(dto.Response{Status: "refresh token is required"})
}
