package http

import (
	"TaskFlow/internal/domain"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"TaskFlow/internal/service"

	"github.com/go-chi/chi/v5"
)

type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

type createProjectReq struct {
	Name string `json:"name"`
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}

	var req createProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, 400, "BAD_REQUEST", "invalid json", nil)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
			[]ErrorDetail{{Field: "name", Message: "is required"}})
		return
	}

	p, err := h.svc.Create(r.Context(), uid, req.Name)
	if err != nil {
		WriteError(w, 500, "INTERNAL", "failed to create project", nil)
		return
	}
	WriteJSON(w, 201, map[string]any{"data": p})
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
				[]ErrorDetail{{Field: "limit", Message: "must be an integer"}})
			return
		}
		limit = n
	}

	var cursor *domain.Cursor
	cAt := r.URL.Query().Get("cursorCreatedAt")
	cID := r.URL.Query().Get("cursorId")
	if cAt != "" || cID != "" {
		// require both for a valid cursor
		if cAt == "" || cID == "" {
			WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
				[]ErrorDetail{{Field: "cursor", Message: "cursorCreatedAt and cursorId must both be provided"}})
			return
		}
		tm, err := time.Parse(time.RFC3339, cAt)
		if err != nil {
			WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
				[]ErrorDetail{{Field: "cursorCreatedAt", Message: "must be RFC3339 timestamp"}})
			return
		}
		cursor = &domain.Cursor{CreatedAt: tm, ID: cID}
	}

	page, err := h.svc.List(r.Context(), uid, limit, cursor)
	if err != nil {
		WriteError(w, 500, "INTERNAL", "failed to list projects", nil)
		return
	}

	resp := map[string]any{"data": page.Items}
	if page.NextCursor != nil {
		resp["meta"] = map[string]any{"nextCursor": page.NextCursor}
	}
	WriteJSON(w, 200, resp)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}

	id := chi.URLParam(r, "id")
	p, err := h.svc.Get(r.Context(), uid, id)
	if err != nil {
		if err == service.ErrNotFound || err == sql.ErrNoRows {
			WriteError(w, 404, "NOT_FOUND", "project not found", nil)
			return
		}
		WriteError(w, 500, "INTERNAL", "failed to get project", nil)
		return
	}
	WriteJSON(w, 200, map[string]any{"data": p})
}

type updateProjectReq struct {
	Name *string `json:"name"`
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}
	id := chi.URLParam(r, "id")

	var req updateProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, 400, "BAD_REQUEST", "invalid json", nil)
		return
	}
	if req.Name == nil {
		WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
			[]ErrorDetail{{Field: "name", Message: "is required"}})
		return
	}
	name := strings.TrimSpace(*req.Name)
	if name == "" {
		WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
			[]ErrorDetail{{Field: "name", Message: "cannot be empty"}})
		return
	}

	p, err := h.svc.UpdateName(r.Context(), uid, id, name)
	if err != nil {
		if err == service.ErrNotFound {
			WriteError(w, 404, "NOT_FOUND", "project not found", nil)
			return
		}
		WriteError(w, 500, "INTERNAL", "failed to update project", nil)
		return
	}
	WriteJSON(w, 200, map[string]any{"data": p})
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}
	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), uid, id); err != nil {
		if err == service.ErrNotFound {
			WriteError(w, 404, "NOT_FOUND", "project not found", nil)
			return
		}
		WriteError(w, 500, "INTERNAL", "failed to delete project", nil)
		return
	}
	w.WriteHeader(204)
}
