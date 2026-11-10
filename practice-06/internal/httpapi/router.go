package httpapi

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
	"time"
)

func BuildRouter(d *gorm.DB) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(5 * time.Second))
	h := Handlers{db: d}
	r.Get("/health", h.Health)
	r.Post("/users", h.CreateUser)
	r.Get("/users", h.ListUsers)
	r.Post("/notes", h.CreateNote)
	r.Get("/notes/{id}", h.GetNoteByID)
	return r
}
