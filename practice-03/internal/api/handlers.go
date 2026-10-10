package api

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"

	//ошибка в инициализации модуля (pracrice)
	"github.com/NikaSof/industrial-programming/pracrice-03/internal/storage"
)

type Handlers struct {
	Store *storage.MemoryStore
}

func NewHandlers(store *storage.MemoryStore) *Handlers {
	return &Handlers{Store: store}
}

func (h *Handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.Store.List()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query != "" {
		query = strings.ToLower(query)
		filtered := make([]storage.Task, 0, len(tasks))

		for _, task := range tasks {
			if strings.Contains(strings.ToLower(task.Title), query) {
				filtered = append(filtered, task)
			}
		}

		tasks = filtered
	}

	JSON(w, http.StatusOK, tasks)
}

type createTaskRequest struct {
	Title string `json:"title"`
}

func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	if contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			BadRequest(w, "Content-Type must be application/json")
			return
		}
	}

	var req createTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid JSON")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		BadRequest(w, "title is required")
		return
	}

	task := h.Store.Create(req.Title)
	JSON(w, http.StatusCreated, task)
}

func (h *Handlers) GetTask(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != "tasks" {
		NotFound(w, "invalid path")
		return
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		BadRequest(w, "invalid id")
		return
	}

	task, err := h.Store.Get(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			NotFound(w, "task not found")
			return
		}

		Internal(w, "unexpected error")
		return
	}

	JSON(w, http.StatusOK, task)
}
