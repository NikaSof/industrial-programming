package api

import (
	"log"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	// Информационные ответы не считаются окончательным статусом.
	if code >= 100 && code < 200 {
		sr.ResponseWriter.WriteHeader(code)
		return
	}

	if sr.status != 0 {
		return
	}

	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(data []byte) (int, error) {
	// Первая запись тела без явного статуса означает 200 OK.
	if sr.status == 0 {
		sr.WriteHeader(http.StatusOK)
	}

	return sr.ResponseWriter.Write(data)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		log.Printf(
			"%s %s %d %v",
			r.Method,
			r.URL.Path,
			status,
			time.Since(start),
		)
	})
}
