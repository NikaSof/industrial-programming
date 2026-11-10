// Local setup reuses the existing practice-05 connection without displaying secrets.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"os"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Println("Setup failed; check local PostgreSQL connection")
		os.Exit(1)
	}
}
func run() error {
	values, e := godotenv.Read("../practice-05/.env")
	if e != nil {
		return e
	}
	cfg, e := pgx.ParseConfig(values["DATABASE_URL"])
	if e != nil {
		return e
	}
	cfg.Database = "postgres"
	d, e := sql.Open("pgx", cfg.ConnString())
	if e != nil {
		return e
	}
	defer d.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var exists bool
	if e = d.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", "pz6_gorm").Scan(&exists); e != nil {
		return e
	}
	if !exists {
		if _, e = d.ExecContext(ctx, "CREATE DATABASE pz6_gorm"); e != nil {
			return e
		}
	}
	cfg.Database = "pz6_gorm"
	if _, e = os.Stat(".env"); os.IsNotExist(e) {
		if e = godotenv.Write(map[string]string{"DB_DSN": cfg.ConnString(), "PORT": "8080"}, ".env"); e != nil {
			return e
		}
	}
	fmt.Println("Database pz6_gorm ready; local .env ready")
	return nil
}
