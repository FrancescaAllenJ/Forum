package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error

	// Open or create the SQLite database file
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}

	// Run the schema.sql to build tables
	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		fmt.Println("Error reading schema file:", err)
		os.Exit(1)
	}

	_, err = DB.Exec(string(schema))
	if err != nil {
		fmt.Println("Error executing schema:", err)
		os.Exit(1)
	}

	fmt.Println("Database initialised successfully.")
}
