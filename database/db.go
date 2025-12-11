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

	// Check if database already exists
	_, err = os.Stat("./forum.db")
	dbExists := err == nil

	// Open SQLite database
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// SQLite stability settings
	DB.SetMaxOpenConns(1)

	// Enable foreign key constraints
	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Run schema only if DB file did not exist yet
	if !dbExists {
		log.Println("📌 Database not found — creating fresh schema...")

		schema, err := os.ReadFile("database/schema.sql")
		if err != nil {
			log.Fatalf("Error reading schema file: %v", err)
		}

		_, err = DB.Exec(string(schema))
		if err != nil {
			log.Fatalf("Error executing schema: %v", err)
		}

		SeedCategories()
		log.Println("📦 Fresh database created successfully.")

	} else {
		log.Println("📦 Database exists — skipping schema creation.")
	}
}

func SeedCategories() {
	categories := []string{
		"General",
		"News",
		"Tech",
		"Advice",
		"Music",
		"Random",
	}

	for _, name := range categories {
		_, err := DB.Exec(`
			INSERT OR IGNORE INTO categories (name)
			VALUES (?)
		`, name)

		if err != nil {
			log.Println("Error inserting category:", name, err)
		}
	}
}
