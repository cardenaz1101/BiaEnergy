package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/biaenergy/backend/internal/domain"
)

const maxBodyBytes = 1 << 20

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeServiceError(w http.ResponseWriter, err error, notFoundMessage string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, notFoundMessage)
	case errors.Is(err, domain.ErrInvalidStatus):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return false
	}
	return true
}
