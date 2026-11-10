package integration_test

import (
	"borovskaya/pr6/internal/db"
	"borovskaya/pr6/internal/httpapi"
	"borovskaya/pr6/internal/models"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRelations(t *testing.T) {
	_ = godotenv.Load()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("DB_DSN required")
	}
	d, e := db.Connect(dsn)
	if e != nil {
		t.Fatal("database unavailable")
	}
	pool, _ := d.DB()
	defer pool.Close()
	if e = d.AutoMigrate(&models.User{}, &models.Note{}, &models.Tag{}); e != nil {
		t.Fatal(e)
	}
	tx := d.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	router := httpapi.BuildRouter(tx)
	request := func(method, path, body string, status int) []byte {
		t.Helper()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w.Body.Bytes()
	}
	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	body := fmt.Sprintf(`{"name":"Тест","email":%q}`, email)
	var u models.User
	json.Unmarshal(request("POST", "/users", body, 201), &u)
	var listed []models.User
	if err := json.Unmarshal(request("GET", "/users", "", 200), &listed); err != nil {
		t.Fatal(err)
	}
	found := false
	for i, item := range listed {
		if i > 0 && listed[i-1].ID >= item.ID {
			t.Fatal("users not ordered by ID")
		}
		if item.ID == u.ID && item.Email == email {
			found = true
		}
	}
	if !found {
		t.Fatal("created user missing from GET /users")
	}

	noteBody := fmt.Sprintf(`{"title":"Заметка","content":"Текст","userId":%d,"tags":["go","go","gorm"]}`, u.ID)
	var n models.Note
	json.Unmarshal(request("POST", "/notes", noteBody, 201), &n)
	if n.User.ID != u.ID || len(n.Tags) != 2 {
		t.Fatalf("relations not loaded: %+v", n)
	}
	var got models.Note
	json.Unmarshal(request("GET", fmt.Sprintf("/notes/%d", n.ID), "", 200), &got)
	if got.Title != "Заметка" || len(got.Tags) != 2 {
		t.Fatal("preload failed")
	}
	request("GET", "/notes/abc", "", 400)
	request("GET", "/notes/9223372036854775807", "", 404)
	request("POST", "/notes", `{"title":"bad","userId":9223372036854775807,"tags":["unused"]}`, 404)
	var count int64
	tx.Model(&models.Tag{}).Where("name = ?", "unused").Count(&count)
	if count != 0 {
		t.Fatal("orphan tag")
	}
	var users []models.User
	if e = tx.Preload("Notes").Where("id = ?", u.ID).Find(&users).Error; e != nil || len(users) != 1 || len(users[0].Notes) != 1 {
		t.Fatal("1:N failed")
	}
	request("POST", "/users", body, 409)
}
