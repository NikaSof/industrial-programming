package notes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Patch("/{id}", h.patch)
	r.Delete("/{id}", h.del)
	return r
}

func reqCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 5*time.Second)
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	// Кодируем до отправки статуса: при ошибке ещё можно вернуть 500.
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("encode response: %v", err)
		code = http.StatusInternalServerError
		data = []byte(`{"error":"internal_server_error"}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write(append(data, '\n'))
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	// Разрешён один JSON-документ, а не несколько объектов подряд.
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	return true
}

func handleRepoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "title_already_exists")
	case errors.Is(err, ErrInvalid):
		writeError(w, http.StatusBadRequest, "invalid_note_data")
	default:
		log.Printf("notes repository: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_server_error")
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Title) == "" {
		writeError(w, http.StatusBadRequest, "title_required")
		return
	}
	ctx, cancel := reqCtx(r)
	defer cancel()
	note, err := h.repo.Create(ctx, in.Title, in.Content)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := reqCtx(r)
	defer cancel()
	note, err := h.repo.ByID(ctx, chi.URLParam(r, "id"))
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit, err := parseIntParam(query.Get("limit"), 20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_limit")
		return
	}
	skip, err := parseIntParam(query.Get("skip"), 0)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_skip")
		return
	}
	ctx, cancel := reqCtx(r)
	defer cancel()
	items, err := h.repo.List(ctx, query.Get("q"), limit, skip)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func parseIntParam(raw string, fallback int64) (int64, error) {
	if raw == "" {
		return fallback, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	// null, как и отсутствующее поле, не задаёт новое строковое значение.
	if in.Title == nil && in.Content == nil {
		writeError(w, http.StatusBadRequest, "no_fields_to_update")
		return
	}
	if in.Title != nil && strings.TrimSpace(*in.Title) == "" {
		writeError(w, http.StatusBadRequest, "title_required")
		return
	}
	ctx, cancel := reqCtx(r)
	defer cancel()
	note, err := h.repo.Update(ctx, chi.URLParam(r, "id"), in.Title, in.Content)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (h *Handler) del(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := reqCtx(r)
	defer cancel()
	if err := h.repo.Delete(ctx, chi.URLParam(r, "id")); err != nil {
		handleRepoError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
