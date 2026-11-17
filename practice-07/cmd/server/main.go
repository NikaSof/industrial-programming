package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"borovskaya/pr7/internal/cache"
)

const cacheTTL = 10 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	c := cache.New(addr)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err := c.Ping(ctx)
	cancel()
	if err != nil {
		return fmt.Errorf("connect to Redis: %w", err)
	}

	server := &http.Server{
		Addr:              "127.0.0.1:" + port,
		Handler:           newMux(c),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("Redis connected: %s", addr)
	log.Printf("Server listening: http://%s", server.Addr)
	return server.ListenAndServe()
}

func newMux(c *cache.Cache) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		if key == "" || value == "" {
			http.Error(w, "key and value required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := c.Set(ctx, key, value, cacheTTL); err != nil {
			cacheError(w, err)
			return
		}
		writeText(w, fmt.Sprintf("OK: %s=%s (TTL 10s)", key, value))
	})

	mux.HandleFunc("GET /get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		value, err := c.Get(ctx, key)
		if errors.Is(err, cache.ErrMiss) {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}
		if err != nil {
			cacheError(w, err)
			return
		}
		writeText(w, fmt.Sprintf("VALUE: %s=%s", key, value))
	})

	mux.HandleFunc("GET /ttl", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		ttl, err := c.TTL(ctx, key)
		if err != nil {
			cacheError(w, err)
			return
		}
		// go-redis возвращает -2 и -1 без умножения на time.Second.
		switch ttl {
		case -2:
			http.Error(w, "key not found", http.StatusNotFound)
		case -1:
			writeText(w, fmt.Sprintf("TTL for %s: no expiration", key))
		default:
			writeText(w, fmt.Sprintf("TTL for %s: %v", key, ttl))
		}
	})
	return mux
}

func writeText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	fmt.Fprintln(w, text)
}

func cacheError(w http.ResponseWriter, err error) {
	log.Printf("Redis error: %v", err)
	http.Error(w, "cache unavailable", http.StatusServiceUnavailable)
}
