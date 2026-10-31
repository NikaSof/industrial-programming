package main

import (
	"borovskaya/pr5/internal/task"
	"context"
	"errors"
	"github.com/joho/godotenv"
	"os"
	"testing"
	"time"
)

func TestPostgresCRUD(t *testing.T) {
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is required")
	}
	db, err := openDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	repo := task.NewPostgresRepo(db)
	created, err := repo.Create(ctx, "Проверка UTF-8")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DELETE FROM tasks WHERE id=$1", created.ID)
	updated, err := repo.Update(ctx, created.ID, "Изменено", true)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Done || updated.Title != "Изменено" {
		t.Fatalf("bad update: %+v", updated)
	}
	found, err := repo.FindByID(ctx, created.ID)
	if err != nil || found.Title != "Изменено" {
		t.Fatalf("find: %v", err)
	}
	list, err := repo.ListDone(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	matched := false
	for _, v := range list {
		if !v.Done {
			t.Fatal("filter mismatch")
		}
		if v.ID == created.ID {
			matched = true
		}
	}
	if !matched {
		t.Fatal("created task missing from filter")
	}
	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, created.ID); !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("expected not found: %v", err)
	}
}
