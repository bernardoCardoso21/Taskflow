package http

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type ErrorBody struct {
	Error struct {
		Code    string        `json:"code"`
		Message string        `json:"message"`
		Details []ErrorDetail `json:"details,omitempty"`
	} `json:"error"`
}

func WriteError(w http.ResponseWriter, status int, code, msg string, details []ErrorDetail) {
	body := ErrorBody{}
	body.Error.Code = code
	body.Error.Message = msg
	body.Error.Details = details
	WriteJSON(w, status, body)
}

func RecovererJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				WriteError(w, 500, "INTERNAL", "internal server error", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
