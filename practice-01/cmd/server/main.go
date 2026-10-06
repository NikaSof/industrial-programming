package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type healthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if _, err := w.Write([]byte("Hello, world!\n")); err != nil {
		log.Printf("Ошибка записи ответа %v", err)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := user{
		ID:   uuid.NewString(),
		Name: "Vanek",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Ошибка записи ответа %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := healthResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Ошибка записи ответа /health: %v", err)
	}
}

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/user", userHandler)
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:              "localhost:" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Сервер запущен: http://%s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
