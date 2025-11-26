package main

import (
	"html/template"
	"log"
	"net/http"

	"forum/database"
	auth "forum/handlers"
	categories "forum/handlers/categories"
	comments "forum/handlers/comments"
	posts "forum/handlers/posts"
	"forum/models"
)

// templates will hold all parsed HTML files.
var templates *template.Template

// HomeData is the structure passed into index.html.
// It contains the logged-in user (if any) and a list of posts.
type HomeData struct {
	User  *auth.SessionUser
	Posts []models.Post
}

func main() {
	// 1. Initialize the database (creates all tables).
	database.InitDB()

	// 2. Load HTML templates.
	loadTemplates()

	// 3. Create router.
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
	mux.HandleFunc("/post", posts.ViewPostHandler)

	// Comment routes
	mux.HandleFunc("/create-comment", comments.CreateCommentHandler)

	// Category filter route
	mux.HandleFunc("/category", categories.ViewCategoryHandler)

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

// homeHandler displays the homepage with posts + categories.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Get logged-in user (nil if not logged in)
	user, _ := auth.GetUserFromRequest(r)

	// Query posts with usernames
	rows, err := database.DB.Query(`
		SELECT posts.id, posts.user_id, users.username, posts.title, posts.content, posts.created_at
		FROM posts
		JOIN users ON posts.user_id = users.id
		ORDER BY posts.created_at DESC
	`)
	if err != nil {
		http.Error(w, "Error loading posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var postsList []models.Post

	// Loop through results
	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.CreatedAt)
		if err != nil {
			http.Error(w, "Error reading posts", http.StatusInternalServerError)
			return
		}

		// Get categories per post
		cats, err := posts.GetCategoriesForPost(p.ID)
		if err == nil {
			p.Categories = cats
		}

		postsList = append(postsList, p)
	}

	// Build data for template
	data := HomeData{
		User:  user,
		Posts: postsList,
	}

	// Render homepage
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
