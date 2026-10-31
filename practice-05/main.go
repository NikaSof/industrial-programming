package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"borovskaya/pr5/internal/task"
	myMW "borovskaya/pr5/pkg/middleware"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("Ошибка чтения .env: %v", err)
			return
		}
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Print("Переменная DATABASE_URL не задана")
		return
	}

	db, err := openDB(dsn)
	if err != nil {
		log.Printf("Ошибка подключения к PostgreSQL: %v", err)
		return
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL")

	repo := task.NewPostgresRepo(db)

	seed := flag.Bool("seed", false, "insert sample tasks and exit")
	poolCheck := flag.Bool("pool-check", false, "measure connection pool")
	flag.Parse()
	if *poolCheck {
		if err := checkPool(db); err != nil {
			log.Print(err)
		}
		return
	}

	if *seed {
		if err := seedTasks(repo); err != nil {
			log.Printf("Ошибка демонстрации транзакции: %v", err)
		}
		return
	}

	handler := task.NewHandler(repo)

	router := chi.NewRouter()

	router.Use(chimw.RequestID)
	router.Use(chimw.Recoverer)
	router.Use(myMW.Logger)
	router.Use(myMW.SimpleCORS)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("write health response: %v", err)
		}
	})

	router.Route("/api", func(api chi.Router) {
		api.Mount("/tasks", handler.Routes())
	})

	addr := "localhost:8080"
	log.Printf("Запуск сервера на http://%s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Printf("Ошибка HTTP-сервера: %v", err)
	}
}

func seedTasks(repo *task.PostgresRepo) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	titles := []string{
		"Подготовить практику 5",
		"Проверить транзакцию",
		"Оформить README",
	}

	if err := repo.CreateMany(ctx, titles); err != nil {
		return err
	}

	log.Printf("Транзакция завершена: добавлено задач — %d", len(titles))
	tasks, err := repo.List(ctx)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		log.Printf("#%d | %s | done=%v", t.ID, t.Title, t.Done)
	}
	if len(tasks) > 0 {
		found, err := repo.FindByID(ctx, tasks[0].ID)
		if err != nil {
			return err
		}
		log.Printf("FindByID: %+v", found)
	}
	return nil
}
