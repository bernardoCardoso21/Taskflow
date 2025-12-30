package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"TaskFlow/internal/domain"
	"TaskFlow/internal/service"

	"github.com/go-chi/chi/v5"
)

type TaskHandler struct {
	svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler { return &TaskHandler{svc: svc} }

type createTaskReq struct {
	Title string `json:"title"`
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}

	projectID := chi.URLParam(r, "projectId")

	var req createTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, 400, "BAD_REQUEST", "invalid json", nil)
		return
	}

	task, err := h.svc.Create(r.Context(), uid, projectID, req.Title)
	if err != nil {
		if err == service.ErrNotFound {
			WriteError(w, 404, "NOT_FOUND", "project not found", nil)
			return
		}
		if strings.Contains(err.Error(), "title") {
			WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
				[]ErrorDetail{{Field: "title", Message: err.Error()}})
			return
		}
		WriteError(w, 500, "INTERNAL", "failed to create task", nil)
		return
	}

	WriteJSON(w, 201, map[string]any{"data": task})
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}

	projectID := strings.TrimSpace(r.URL.Query().Get("projectId"))
	if projectID == "" {
		WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
			[]ErrorDetail{{Field: "projectId", Message: "is required"}})
		return
	}

	var completed *bool
	if v := r.URL.Query().Get("completed"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
				[]ErrorDetail{{Field: "completed", Message: "must be true or false"}})
			return
		}
		completed = &b
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

	page, err := h.svc.List(r.Context(), uid, projectID, completed, limit, cursor)
	if err != nil {
		WriteError(w, 500, "INTERNAL", "failed to list tasks", nil)
		return
	}

	resp := map[string]any{"data": page.Items}
	if page.NextCursor != nil {
		resp["meta"] = map[string]any{"nextCursor": page.NextCursor}
	}
	WriteJSON(w, 200, resp)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}

	id := chi.URLParam(r, "id")
	task, err := h.svc.Get(r.Context(), uid, id)
	if err != nil {
		if err == service.ErrNotFound {
			WriteError(w, 404, "NOT_FOUND", "task not found", nil)
			return
		}
		WriteError(w, 500, "INTERNAL", "failed to get task", nil)
		return
	}
	WriteJSON(w, 200, map[string]any{"data": task})
}

type updateTaskReq struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}

	id := chi.URLParam(r, "id")

	var req updateTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, 400, "BAD_REQUEST", "invalid json", nil)
		return
	}
	if req.Title == nil && req.Completed == nil {
		WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
			[]ErrorDetail{{Field: "body", Message: "must include at least one of: title, completed"}})
		return
	}

	task, err := h.svc.Update(r.Context(), uid, id, req.Title, req.Completed)
	if err != nil {
		if err == service.ErrNotFound {
			WriteError(w, 404, "NOT_FOUND", "task not found", nil)
			return
		}
		if strings.Contains(err.Error(), "title") {
			WriteError(w, 422, "VALIDATION_ERROR", "invalid request",
				[]ErrorDetail{{Field: "title", Message: err.Error()}})
			return
		}
		WriteError(w, 500, "INTERNAL", "failed to update task", nil)
		return
	}

	WriteJSON(w, 200, map[string]any{"data": task})
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		WriteError(w, 401, "UNAUTHORIZED", "missing user", nil)
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), uid, id); err != nil {
		if err == service.ErrNotFound {
			WriteError(w, 404, "NOT_FOUND", "task not found", nil)
			return
		}
		WriteError(w, 500, "INTERNAL", "failed to delete task", nil)
		return
	}
	w.WriteHeader(204)
}
