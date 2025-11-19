package handlers

import (
	"avito-assignment-2025-autumn/internal/delivery/http/dto"
	"bytes"
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, payload interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(buf.Bytes())
	return err
}

func writeAPIError(w http.ResponseWriter, status int, code dto.ErrorCode, message string, info ...string) error {
	apiErr := dto.APIError{
		Code:    code,
		Message: message,
	}

	if len(info) > 0 {
		apiErr.Info = info[0]
	}

	return writeJSON(w, status, dto.ErrorResponse{Error: apiErr})
}

func writeInternalServerError(w http.ResponseWriter, err error) {
	var info string
	if err != nil {
		info = err.Error()
	}

	if writeErr := writeAPIError(w, http.StatusInternalServerError, dto.ErrorCodeInternal, "internal server error", info); writeErr != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
