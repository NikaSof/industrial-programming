package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"borovskaya/pr8/internal/db"
	"borovskaya/pr8/internal/notes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return errors.New("MONGO_URI is required: configure .env or environment")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	deps, err := db.ConnectMongo(ctx, uri, getenv("MONGO_DB", "pz8"))
	if err != nil {
		return err
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := deps.Client.Disconnect(cleanupCtx); err != nil {
			log.Printf("disconnect MongoDB: %v", err)
		}
	}()
	repo, err := notes.NewRepo(ctx, deps.Database)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr: getenv("HTTP_ADDR", "127.0.0.1:8080"), Handler: newRouter(repo),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("MongoDB connected: database=%s", deps.Database.Name())
		log.Printf("Listening on http://%s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()
	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		// Даём текущим запросам завершиться до закрытия клиента MongoDB.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		return nil
	}
}

func newRouter(repo *notes.Repo) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Проверка HTTP-процесса; доступность БД проверяется при старте.
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	r.Mount("/api/v1/notes", notes.NewHandler(repo).Routes())
	return r
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
