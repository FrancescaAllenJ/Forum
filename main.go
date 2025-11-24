package main

import (
	"html/template"
	"log"
	"net/http"

	"forum/database"
	auth "forum/handlers"
)

var templates *template.Template

// HomeData is the data passed to index.html
type HomeData struct {
	User *auth.SessionUser
}

func main() {
	// 1. Initialize the database
	database.InitDB()

	// 2. Preload ALL templates
	loadTemplates()

	// 3. Setup routing
	mux := http.NewServeMux()

	// Home
	mux.HandleFunc("/", homeHandler)

	// Registration route
	mux.HandleFunc("/register", auth.RegisterHandler)

	// Login route
	mux.HandleFunc("/login", auth.LoginHandler)

	// Logout route
	mux.HandleFunc("/logout", auth.LogoutHandler)

	// Serve static files
	static := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", static))

	// Start server
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
	// Try to get logged-in user
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		log.Println("Error getting user from session:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	data := HomeData{
		User: user,
	}

	err = templates.ExecuteTemplate(w, "index.html", data)
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
