package main

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"borovskaya/pr8/internal/db"
	"borovskaya/pr8/internal/notes"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNotesAPI(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("set MONGO_TEST_URI to run integration tests with MongoDB")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	// Уникальное имя исключает изменение рабочей базы pz8.
	deps, err := db.ConnectMongo(ctx, uri, "pz8_test_"+bson.NewObjectID().Hex())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := deps.Database.Drop(cleanupCtx); err != nil {
			t.Errorf("drop test database: %v", err)
		}
		if err := deps.Client.Disconnect(cleanupCtx); err != nil {
			t.Errorf("disconnect: %v", err)
		}
	})
	repo, err := notes.NewRepo(ctx, deps.Database)
	if err != nil {
		t.Fatal(err)
	}
	router := newRouter(repo)
	request := func(method, path, body string, status int) []byte {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body)).WithContext(ctx)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, w.Code, status, w.Body.String())
		}
		return w.Body.Bytes()
	}
	decodeNote := func(data []byte) notes.Note {
		t.Helper()
		var note notes.Note
		if err := json.Unmarshal(data, &note); err != nil {
			t.Fatal(err)
		}
		return note
	}
	const base = "/api/v1/notes"
	request("GET", "/health", "", 200)
	if got := strings.TrimSpace(string(request("GET", base, "", 200))); got != "[]" {
		t.Fatalf("empty list: %s", got)
	}
	first := decodeNote(request("POST", base, `{"title":"Первая Go","content":"Текст заметки"}`, 201))
	if first.ID.IsZero() || first.CreatedAt.IsZero() || first.Title != "Первая Go" {
		t.Fatalf("invalid created note: %+v", first)
	}
	path := base + "/" + first.ID.Hex()
	got := decodeNote(request("GET", path, "", 200))
	if got.Content != first.Content || got.ID != first.ID {
		t.Fatal("created note differs from stored note")
	}
	request("POST", base, `{"title":"Первая Go"}`, 409)
	second := decodeNote(request("POST", base+"/", `{"title":"Вторая Mongo","content":"Другой текст"}`, 201))
	var items []notes.Note
	if err := json.Unmarshal(request("GET", base+"?q=%20GO&limit=1&skip=0", "", 200), &items); err != nil || len(items) != 1 || items[0].ID != first.ID {
		t.Fatalf("search failed: %v, %+v", err, items)
	}
	if err := json.Unmarshal(request("GET", base+"?limit=1&skip=1", "", 200), &items); err != nil || len(items) != 1 || items[0].ID != first.ID {
		t.Fatalf("pagination failed: %v, %+v", err, items)
	}
	if got := strings.TrimSpace(string(request("GET", base+"?q=.", "", 200))); got != "[]" {
		t.Fatal("search text was treated as regular expression")
	}
	updated := decodeNote(request("PATCH", path, `{"content":"Обновлено"}`, 200))
	if updated.Title != first.Title || updated.Content != "Обновлено" || !updated.CreatedAt.Equal(first.CreatedAt) || updated.UpdatedAt.Before(first.UpdatedAt) {
		t.Fatalf("partial update failed: %+v", updated)
	}
	cleared := decodeNote(request("PATCH", path, `{"content":""}`, 200))
	if cleared.Content != "" || cleared.Title != first.Title {
		t.Fatal("clear content failed")
	}
	request("PATCH", base+"/"+second.ID.Hex(), `{"title":"Первая Go"}`, 409)
	for _, body := range []string{`{}`, `{"title":" "}`, `{"content":null}`, `{"unknown":1}`, `{} {}`} {
		request("PATCH", path, body, 400)
	}
	request("POST", base, `{"title":" "}`, 400)
	request("POST", base, `{"title":"X","typo":true}`, 400)
	request("GET", base+"?limit=abc", "", 400)
	request("GET", base+"/invalid", "", 404)
	request("GET", base+"/"+bson.NewObjectID().Hex(), "", 404)
	if body := request("DELETE", path, "", 204); len(body) != 0 {
		t.Fatal("204 response must be empty")
	}
	request("GET", path, "", 404)
	request("DELETE", path, "", 404)
}
