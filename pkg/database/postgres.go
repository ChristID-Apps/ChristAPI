package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

_ 	"github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresConnection() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// connection pool config
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 5)

	// test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}