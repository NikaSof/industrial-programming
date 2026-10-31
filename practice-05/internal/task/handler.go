package task

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo *PostgresRepo
}

type updateReq struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Вспомогательные функции

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		httpError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}

	return id, true
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("encode JSON: %v", err)
		code = http.StatusInternalServerError
		data = []byte(`{"error":"internal server error"}`)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Printf("write JSON: %v", err)
	}
}

func httpError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{
		"error": msg,
	})
}

// Основные функции

func NewHandler(repo *PostgresRepo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)

	return r
}

// list
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var (
		tasks []Task
		err   error
	)

	query := r.URL.Query()

	if query.Has("done") {
		raw := query.Get("done")
		if raw != "true" && raw != "false" {
			httpError(w, http.StatusBadRequest, "done must be true or false")
			return
		}

		tasks, err = h.repo.ListDone(ctx, raw == "true")
	} else {
		tasks, err = h.repo.List(ctx)
	}

	if err != nil {
		handleRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	task, err := h.repo.Get(ctx, id)
	if err != nil {
		handleRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// create
type createReq struct {
	Title string `json:"title"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		httpError(w, http.StatusBadRequest, "title is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	task, err := h.repo.Create(ctx, req.Title)
	if err != nil {
		handleRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

// update
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		httpError(w, http.StatusBadRequest, "title is required")
		return
	}
	task, err := h.repo.Update(ctx, id, req.Title, req.Done)
	if err != nil {
		handleRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// delete
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.repo.Delete(ctx, id); err != nil {
		handleRepoError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleRepoError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		httpError(w, http.StatusNotFound, "task not found")
		return
	}

	log.Printf("repository error: %v", err)
	httpError(w, http.StatusInternalServerError, "internal server error")
}
