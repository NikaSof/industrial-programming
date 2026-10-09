package app

import (
	"fmt"
	"net/http"

	"github.com/NikaSof/indastrial-programming/practice-02/internal/app/handlers"
	"github.com/NikaSof/indastrial-programming/practice-02/utils"
)

// type pingResp struct {
// 	Status string `json:"status"`
// 	Time   string `json:"time"`
// }

func Run() {
	mux := http.NewServeMux()
	handler := withRequestID(mux)

	mux.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			utils.LogRequest(r)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			if _, err := fmt.Fprintln(w, "Hellow, Go project structure!"); err != nil {
				utils.LogError("write response: " + err.Error())
			}
		})

	// mux.HandleFunc("/ping",
	// 	func(w http.ResponseWriter, r *http.Request) {
	// 		utils.LogRequest(r)
	// 		w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 		response := pingResp{
	// 			Status: "ok",
	// 			Time:   time.Now().UTC().Format(time.RFC3339),
	// 		}

	// 		if err := json.NewEncoder(w).Encode(response); err != nil {
	// 			utils.LogError("write JSON:" + err.Error())
	// 		}
	// 	})

	mux.HandleFunc("/ping", handlers.Ping)

	mux.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)
		utils.WriteErr(w, http.StatusBadRequest, "bad_request_example")
	})

	utils.LogInfo("Server is starting on :8080")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		utils.LogError("server error: " + err.Error())
	}
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")

		if id == "" {
			var err error
			id, err = utils.NewID16()
			if err != nil {
				utils.LogError("generate request ID: " + err.Error())
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r)
	})

}
