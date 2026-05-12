package mapper_err

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
	default:
		writeJSON(w, http.StatusInternalServerError, dto.Error("internal server error"))
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
