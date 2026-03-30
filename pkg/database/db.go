package database

import (
	"database/sql"
	"log"
)

var DB *sql.DB

func Connect() {
	conn, err := NewPostgresConnection()
	if err != nil {
		log.Fatal("❌ Failed to connect database:", err)
	}

	DB = conn

	log.Println("✅ PostgreSQL Connected")
}