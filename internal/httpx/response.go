package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorBody{Error: ErrorDetail{Code: code, Message: message}})
}

// WriteInternalError logs the underlying error server-side and returns a
// generic 500 response, so internal failure details are never exposed to
// the client.
func WriteInternalError(w http.ResponseWriter, err error) {
	slog.Error("internal server error", "error", err)
	WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
