package httputil

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/derletzte256/avito-assignment-2025-autumn/internal/entity"
)

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(buf.Bytes())
	return err
}

func WriteAPIError(w http.ResponseWriter, status int, code entity.ErrorCode, message string, info ...string) error {
	apiErr := entity.APIError{
		Code:    code,
		Message: message,
	}

	if len(info) > 0 {
		apiErr.Info = info[0]
	}

	if writeErr := WriteJSON(w, status, entity.ErrorResponse{Error: apiErr}); writeErr != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return writeErr
	}

	return nil
}

func WriteInternalServerError(w http.ResponseWriter, err error) {
	var info string
	if err != nil {
		info = err.Error()
	}

	if writeErr := WriteAPIError(w, http.StatusInternalServerError, entity.ErrorCodeInternal, "internal server error", info); writeErr != nil {
		return
	}
}

func ReadJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(dst)
}
