package main

import (
	"html/template"
	"log"
	"net/http"

	"forum/database"
	auth "forum/handlers"
	posts "forum/handlers/posts"
)

var templates *template.Template

// HomeData holds the data sent to index.html
type HomeData struct {
	User  *auth.SessionUser
	Posts []Post
}

// Post structure used for displaying posts on the homepage
type Post struct {
	ID        int
	UserID    int
	Title     string
	Content   string
	CreatedAt string
}

func main() {
	// 1. Initialize the database
	database.InitDB()

	// 2. Load all templates
	loadTemplates()

	// 3. Routing
	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/register", auth.RegisterHandler)
	mux.HandleFunc("/login", auth.LoginHandler)
	mux.HandleFunc("/logout", auth.LogoutHandler)

	// NEW — Create Post Route
	mux.HandleFunc("/create-post", posts.CreatePostHandler)

	// Static files
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
	user, _ := auth.GetUserFromRequest(r)

	// Load posts from DB
	rows, err := database.DB.Query(`
		SELECT id, user_id, title, content, created_at
		FROM posts
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, "Error loading posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.CreatedAt)
		posts = append(posts, p)
	}

	data := HomeData{
		User:  user,
		Posts: posts,
	}

	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
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
