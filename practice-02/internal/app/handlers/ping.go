package handlers

import (
	"net/http"
	"time"

	"github.com/NikaSof/indastrial-programming/practice-02/utils"
)

type pingResp struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func Ping(w http.ResponseWriter, r *http.Request) {
	utils.LogRequest(r)

	response := pingResp{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
