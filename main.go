package main

import (
	"html/template"
	"log"
	"net/http"

	"forum/database"
)

var templates *template.Template

func main() {
	// 1. Initialize the database
	database.InitDB()

	// 2. Preload ALL templates
	loadTemplates()

	// 3. Setup routing
	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)

	// Serve static files
	static := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", static))

	log.Println("Server running at http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func loadTemplates() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// WHAT A WEB SERVER ACTUALLY IS

// A web server is a program that:

// Listens for incoming network requests

// Understands HTTP

// Decides which function to run for each URL

// Sends responses back to the browser

// Your Go code is doing all four.

// net/http - STAR OF THE SHOW
// create a server, listen on a port, read requests, send responses, handle cookies, manage HTTP methods (GET, POST)
// define handlers, This is why Go is famous for web development — the HTTP server is built into the language.
