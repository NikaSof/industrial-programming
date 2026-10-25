package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"borovskaya/pr4/internal/task"
	myMW "borovskaya/pr4/pkg/middleware"
)

func main() {
	repo := task.NewRepo()
	handler := task.NewHandler(repo)

	router := chi.NewRouter()

	router.Use(chimw.RequestID)
	router.Use(chimw.Recoverer)
	router.Use(myMW.Logger)
	router.Use(myMW.SimpleCORS)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "test/plain; charset=utf-8")

		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("write health response: %v", err)
		}
	})

	router.Route("/api", func(api chi.Router) {
		api.Mount("/tasks", handler.Routes())
	})

	addr := "localhost:8080"
	log.Printf("Запуск сервера на http://%s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
