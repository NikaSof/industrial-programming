package main

import (
	"borovskaya/pr6/internal/db"
	"borovskaya/pr6/internal/httpapi"
	"borovskaya/pr6/internal/models"
	"context"
	"errors"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)

func run() error {
	if e := godotenv.Load(); e != nil && !errors.Is(e, os.ErrNotExist) {
		return errors.New("cannot read .env")
	}
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return errors.New("DB_DSN is empty")
	}
	d, e := db.Connect(dsn)
	if e != nil {
		return errors.New("PostgreSQL connection failed; check DB_DSN and server")
	}
	pool, _ := d.DB()
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if e = d.WithContext(ctx).AutoMigrate(&models.User{}, &models.Note{}, &models.Tag{}); e != nil {
		return e
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := http.Server{Addr: "localhost:" + port, Handler: httpapi.BuildRouter(d), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("PostgreSQL connected, migrations complete; http://%s", server.Addr)
	return server.ListenAndServe()
}
func main() {
	if e := run(); e != nil {
		log.Print(e)
		os.Exit(1)
	}
}
