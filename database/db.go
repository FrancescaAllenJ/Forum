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

	// Enable foreign key constraints (VERY IMPORTANT)
	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Load schema file (create tables)
	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		log.Fatalf("Error reading schema file: %v", err)
	}

	_, err = DB.Exec(string(schema))
	if err != nil {
		log.Fatalf("Error executing schema: %v", err)
	}

	// --- NEW: Seed initial categories ---
	SeedCategories()

	log.Println("📦 Database initialised successfully.")
}

// SeedCategories inserts default category names if they don't already exist.
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
		// INSERT OR IGNORE ensures no duplicate categories are created.
		_, err := DB.Exec(`
			INSERT OR IGNORE INTO categories (name)
			VALUES (?)
		`, name)

		if err != nil {
			log.Println("Error inserting category:", name, err)
		}
	}
}
