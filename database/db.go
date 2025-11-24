package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error

	// Open SQLite database file
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Ensure SQLite behaves correctly (only one connection allowed)
	DB.SetMaxOpenConns(1)

	// Force SQLite to honour foreign key constraints
	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Load schema file
	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		log.Fatalf("Error reading schema file: %v", err)
	}

	// Execute schema (create tables)
	_, err = DB.Exec(string(schema))
	if err != nil {
		log.Fatalf("Error executing schema: %v", err)
	}

	log.Println("📦 Database initialised successfully.")
}
