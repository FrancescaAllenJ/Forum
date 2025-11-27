package main

import (
	"html/template"
	"log"
	"net/http"

	"forum/database"
	auth "forum/handlers"
	categories "forum/handlers/categories"
	comments "forum/handlers/comments"
	likes "forum/handlers/likes"
	posts "forum/handlers/posts"
	"forum/models"
)

var templates *template.Template

// Data passed to index.html
type HomeData struct {
	User  *auth.SessionUser
	Posts []models.Post
}

func main() {
	// --------------------------
	// INITIALIZATION
	// --------------------------
	database.InitDB()
	loadTemplates()

	mux := http.NewServeMux()

	// --------------------------
	// MAIN PAGES
	// --------------------------
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/register", auth.RegisterHandler)
	mux.HandleFunc("/login", auth.LoginHandler)
	mux.HandleFunc("/logout", auth.LogoutHandler)

	// --------------------------
	// POSTS
	// --------------------------
	mux.HandleFunc("/create-post", posts.CreatePostHandler)
	mux.HandleFunc("/post", posts.ViewPostHandler)

	// NEW — filter routes
	mux.HandleFunc("/my-posts", posts.MyPostsHandler)
	mux.HandleFunc("/liked-posts", posts.LikedPostsHandler)

	// --------------------------
	// COMMENTS
	// --------------------------
	mux.HandleFunc("/create-comment", comments.CreateCommentHandler)

	// --------------------------
	// CATEGORIES
	// --------------------------
	mux.HandleFunc("/category", categories.ViewCategoryHandler)

	// --------------------------
	// LIKES
	// --------------------------
	mux.HandleFunc("/like", likes.LikeHandler)

	// --------------------------
	// STATIC FILES
	// --------------------------
	static := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", static))

	// --------------------------
	// CUSTOM 404 HANDLER
	// --------------------------
	mux.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		templates.ExecuteTemplate(w, "error_404.html", nil)
	})

	log.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// --------------------------------------
// TEMPLATE LOADER
// --------------------------------------
func loadTemplates() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
}

// --------------------------------------
// HOMEPAGE HANDLER (fixed + stable)
// --------------------------------------
func homeHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Home handler started")

	user, _ := auth.GetUserFromRequest(r)

	// STEP 1 — Query posts (raw)
	log.Println("Running post query…")
	rows, err := database.DB.Query(`
		SELECT posts.id,
		       posts.user_id,
		       users.username,
		       posts.title,
		       posts.content,
		       strftime('%Y-%m-%d %H:%M:%S', posts.created_at) AS created_at
		FROM posts
		JOIN users ON posts.user_id = users.id
		ORDER BY posts.created_at DESC
	`)
	if err != nil {
		log.Println("Post query error:", err)
		http.Error(w, "Error loading posts", http.StatusInternalServerError)
		return
	}

	type rawPost struct {
		ID        int
		UserID    int
		Username  string
		Title     string
		Content   string
		CreatedAt string
	}

	var rawPosts []rawPost

	for rows.Next() {
		var rp rawPost
		if err := rows.Scan(&rp.ID, &rp.UserID, &rp.Username, &rp.Title, &rp.Content, &rp.CreatedAt); err != nil {
			log.Println("SCAN ERROR:", err)
			continue
		}
		rawPosts = append(rawPosts, rp)
	}

	rows.Close() // essential for SQLite

	// STEP 2 — Enrich with categories + likes
	var postsList []models.Post

	for _, rp := range rawPosts {
		p := models.Post{
			ID:        rp.ID,
			UserID:    rp.UserID,
			Username:  rp.Username,
			Title:     rp.Title,
			Content:   rp.Content,
			CreatedAt: rp.CreatedAt,
		}

		// categories
		cats, err := posts.GetCategoriesForPost(p.ID)
		if err == nil {
			p.Categories = cats
		}

		// likes/dislikes
		lc, dc := likes.CountPostLikes(p.ID)
		p.Likes = lc
		p.Dislikes = dc

		postsList = append(postsList, p)
	}

	// STEP 3 — Render homepage
	data := HomeData{
		User:  user,
		Posts: postsList,
	}

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Println("Template error:", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
