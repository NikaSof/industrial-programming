package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Ошибка преобразования в JSON: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		if _, writeErr := w.Write([]byte(`{"error":"internal server error"}`)); writeErr != nil {
			log.Printf("Ошибка записи ответа: %v", writeErr)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Printf("Ошибка записи ответа: %v", err)
	}
}

func BadRequest(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusBadRequest, ErrorResponse{Error: msg})
}

func NotFound(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusNotFound, ErrorResponse{Error: msg})
}

func Internal(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusInternalServerError, ErrorResponse{Error: msg})
}
