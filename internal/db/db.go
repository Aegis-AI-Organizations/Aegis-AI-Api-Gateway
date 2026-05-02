package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	if host == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("missing one or more required database environment variables")
	}

	escapedPassword := url.QueryEscape(password)
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, escapedPassword, host, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
