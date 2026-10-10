package main

import (
	"log"
	"net/http"
	"os"

	//ошибка в инициализации модуля (pracrice)
	"github.com/NikaSof/industrial-programming/pracrice-03/internal/api"
	"github.com/NikaSof/industrial-programming/pracrice-03/internal/storage"
)

func main() {
	store := storage.NewMemoryStore()
	handlers := api.NewHandlers(store)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		api.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	mux.HandleFunc("GET /tasks", handlers.ListTasks)
	mux.HandleFunc("POST /tasks", handlers.CreateTask)
	mux.HandleFunc("GET /tasks/", handlers.GetTask)

	handler := api.Logging(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := "localhost:" + port
	log.Println("Запуск сервера на", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
