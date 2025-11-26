package main

import (
	"html/template"
	"log"
	"net/http"

	"forum/database"
	auth "forum/handlers"
	comments "forum/handlers/comments"
	posts "forum/handlers/posts"
)

// templates will hold all parsed HTML files.
var templates *template.Template

// HomeData is the structure passed into index.html.
// It contains the logged-in user (if any) and a list of posts.
type HomeData struct {
	User  *auth.SessionUser
	Posts []Post
}

// Post represents a single forum post on the homepage.
type Post struct {
	ID        int
	UserID    int
	Title     string
	Content   string
	CreatedAt string
}

func main() {
	// 1. Initialize database and create all tables.
	database.InitDB()

	// 2. Load all HTML templates at startup.
	loadTemplates()

	// 3. Create a custom HTTP router.
	mux := http.NewServeMux()

	// --------------------------
	// ROUTES
	// --------------------------

	// Homepage
	mux.HandleFunc("/", homeHandler)

	// Auth routes
	mux.HandleFunc("/register", auth.RegisterHandler)
	mux.HandleFunc("/login", auth.LoginHandler)
	mux.HandleFunc("/logout", auth.LogoutHandler)

	// Post routes
	mux.HandleFunc("/create-post", posts.CreatePostHandler)
	mux.HandleFunc("/post", posts.ViewPostHandler) // NEW: single post page

	// Comment routes
	mux.HandleFunc("/create-comment", comments.CreateCommentHandler) // NEW

	// --------------------------
	// STATIC FILES
	// --------------------------

	static := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", static))

	// --------------------------
	// START SERVER
	// --------------------------

	log.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// loadTemplates parses all HTML files into the templates variable.
func loadTemplates() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
}

// homeHandler handles GET "/" and displays the homepage with posts.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get logged-in user (may be nil if not logged in)
	user, _ := auth.GetUserFromRequest(r)

	// 2. Query all posts
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

	// 3. Loop over each row returned from the database
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.CreatedAt); err != nil {
			http.Error(w, "Error reading posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, p)
	}

	// 4. Build data structure to send to index.html
	data := HomeData{
		User:  user,
		Posts: posts,
	}

	// 5. Render the homepage template
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
