package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"treuepunkte/internal/domain"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func mapErrorToStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest

	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrAccrueNotFound),
		errors.Is(err, domain.ErrRedeemNotFound):
		return http.StatusNotFound

	case errors.Is(err, domain.ErrConflict),
		errors.Is(err, domain.ErrInsufficientActivePoints),
		errors.Is(err, domain.ErrTransactionNotPending):
		return http.StatusConflict

	default:
		return http.StatusInternalServerError
	}
}

func writeSuccess(w http.ResponseWriter, created bool) {
	w.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusOK
	statusText := "ok"

	if created {
		statusCode = http.StatusCreated
		statusText = "created"
	}

	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": statusText,
	})
}

func writeMappedError(w http.ResponseWriter, err error) {
	writeError(w, mapErrorToStatus(err), err.Error())
}
