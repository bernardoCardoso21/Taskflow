package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"TaskFlow/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, 400, "BAD_REQUEST", "invalid json", nil)
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
			[]ErrorDetail{{Field: "email", Message: "must be a valid email"}})
		return
	}
	if len(req.Password) < 8 {
		WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
			[]ErrorDetail{{Field: "password", Message: "must be at least 8 chars"}})
		return
	}

	id, err := h.svc.Register(req.Email, req.Password)
	if err != nil {
		WriteError(w, 409, "CONFLICT", "email already registered (or other conflict)", nil)
		return
	}
	WriteJSON(w, 201, map[string]any{"data": map[string]any{"id": id}})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, 400, "BAD_REQUEST", "invalid json", nil)
		return
	}
	token, err := h.svc.Login(strings.TrimSpace(req.Email), req.Password)
	if err != nil {
		WriteError(w, 401, "UNAUTHORIZED", "invalid credentials", nil)
		return
	}
	WriteJSON(w, 200, map[string]any{"data": map[string]any{"accessToken": token}})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}
	WriteJSON(w, 200, map[string]any{"data": map[string]any{"id": uid}})
}
