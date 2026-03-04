package handler

import (
	"encoding/json"
	"kate/services/tasks/internal/client"
	"kate/services/tasks/internal/service"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskHandler struct {
	taskSvc  *service.TaskService
	authGrpc *client.AuthGrpcClient
}

func NewTaskHandler(ts *service.TaskService, ag *client.AuthGrpcClient) *TaskHandler {
	return &TaskHandler{taskSvc: ts, authGrpc: ag}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func (h *TaskHandler) verifyToken(r *http.Request) (bool, string, int) {
	token := extractToken(r)
	if token == "" {
		return false, "", http.StatusUnauthorized
	}

	valid, subject, err := h.authGrpc.VerifyToken(r.Context(), token)
	if err != nil {
		log.Printf("VerifyToken error: %v", err)

		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unavailable, codes.DeadlineExceeded:
				return false, "", http.StatusServiceUnavailable // 503
			}
		}

		if strings.Contains(err.Error(), "unavailable") ||
			strings.Contains(err.Error(), "timeout") ||
			strings.Contains(err.Error(), "connection refused") {
			return false, "", http.StatusServiceUnavailable // 503
		}

		return false, "", http.StatusInternalServerError // 500
	}

	if !valid {
		return false, "", http.StatusUnauthorized // 401
	}

	return true, subject, http.StatusOK
}

func (h *TaskHandler) handleError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// CreateTask — создание задачи
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.handleError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	valid, _, statusCode := h.verifyToken(r)
	if !valid {
		h.handleError(w, statusCode, "unauthorized")
		return
	}

	var task service.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.handleError(w, http.StatusBadRequest, "invalid json")
		return
	}

	created := h.taskSvc.Create(task)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// GetAllTasks — список задач
func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.handleError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	valid, _, statusCode := h.verifyToken(r)
	if !valid {
		h.handleError(w, statusCode, "unauthorized")
		return
	}

	tasks := h.taskSvc.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTaskByID — задача по ID
func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.handleError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	valid, _, statusCode := h.verifyToken(r)
	if !valid {
		h.handleError(w, statusCode, "unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if id == "" {
		h.handleError(w, http.StatusBadRequest, "missing id")
		return
	}

	task, ok := h.taskSvc.GetByID(id)
	if !ok {
		h.handleError(w, http.StatusNotFound, "task not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTask — обновление задачи
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		h.handleError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	valid, _, statusCode := h.verifyToken(r)
	if !valid {
		h.handleError(w, statusCode, "unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if id == "" {
		h.handleError(w, http.StatusBadRequest, "missing id")
		return
	}

	var updates service.Task
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.handleError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updated, ok := h.taskSvc.Update(id, updates)
	if !ok {
		h.handleError(w, http.StatusNotFound, "task not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DeleteTask — удаление задачи
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.handleError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	valid, _, statusCode := h.verifyToken(r)
	if !valid {
		h.handleError(w, statusCode, "unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if id == "" {
		h.handleError(w, http.StatusBadRequest, "missing id")
		return
	}

	ok := h.taskSvc.Delete(id)
	if !ok {
		h.handleError(w, http.StatusNotFound, "task not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
