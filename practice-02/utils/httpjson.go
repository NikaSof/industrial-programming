package utils

import (
	"encoding/json"
	"net/http"
)

type JSONerror struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, code int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		LogError("encode JSON: " + err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if _, err := w.Write(append(data, '\n')); err != nil {
		LogError("write JSON: " + err.Error())
	}
}

func WriteErr(w http.ResponseWriter, code int, msg string) {
	WriteJSON(w, code, JSONerror{Error: msg})
}
