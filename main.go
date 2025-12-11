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

// ------------------------------------------------------------
// DATA STRUCTS
// ------------------------------------------------------------
type HomeData struct {
	User       *auth.SessionUser
	Posts      []models.Post
	Categories []models.Category
}

// ------------------------------------------------------------
// STATUS RECORDER (detects when a route writes nothing → 404)
// ------------------------------------------------------------
type statusRecorder struct {
	http.ResponseWriter
	written bool
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.written = true
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	sr.written = true
	return sr.ResponseWriter.Write(b)
}

// ------------------------------------------------------------
// MAIN
// ------------------------------------------------------------
func main() {
	database.InitDB()
	loadTemplates()

	mux := http.NewServeMux()

	// AUTH
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/register", auth.RegisterHandler)
	mux.HandleFunc("/login", auth.LoginHandler)
	mux.HandleFunc("/logout", auth.LogoutHandler)

	// POSTS
	mux.HandleFunc("/create-post", posts.CreatePostHandler)
	mux.HandleFunc("/post", posts.ViewPostHandler)
	mux.HandleFunc("/my-posts", posts.MyPostsHandler)
	mux.HandleFunc("/liked-posts", posts.LikedPostsHandler)

	// COMMENTS
	mux.HandleFunc("/create-comment", comments.CreateCommentHandler)

	// CATEGORIES
	mux.HandleFunc("/category", categories.ViewCategoryHandler)

	// LIKES
	mux.HandleFunc("/like", likes.LikeHandler)

	// STATIC FILES
	static := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", static))

	// --------------------------------------------------------
	// WRAPPER (Panic → 500, No output → 404)
	// --------------------------------------------------------
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// PANIC PROTECTION (500)
		defer func() {
			if rec := recover(); rec != nil {
				log.Println("Recovered panic:", rec)
				render500(w)
			}
		}()

		// Record output to detect unhandled routes
		rec := &statusRecorder{ResponseWriter: w}

		mux.ServeHTTP(rec, r)

		// No handler wrote → Unknown route → 404
		if !rec.written {
			w.WriteHeader(http.StatusNotFound)
			templates.ExecuteTemplate(w, "error_404.html", nil)
		}
	})

	log.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", wrapped); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// ------------------------------------------------------------
// TEMPLATE LOADER
// ------------------------------------------------------------
func loadTemplates() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
}

// ------------------------------------------------------------
// ERROR 500 PAGE
// ------------------------------------------------------------
func render500(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	templates.ExecuteTemplate(w, "error_500.html", nil)
}

// ------------------------------------------------------------
// HOME HANDLER
// ------------------------------------------------------------
func homeHandler(w http.ResponseWriter, r *http.Request) {

	// MUST match exactly `/`
	if r.URL.Path != "/" {
		return
	}

	user, _ := auth.GetUserFromRequest(r)

	// --------------------------------------------------------
	// LOAD POSTS
	// --------------------------------------------------------
	rows, err := database.DB.Query(`
		SELECT posts.id,
		       posts.user_id,
		       users.username,
		       posts.title,
		       posts.content,
		       strftime('%Y-%m-%d %H:%M:%S', posts.created_at)
		FROM posts
		JOIN users ON posts.user_id = users.id
		ORDER BY posts.created_at DESC
	`)
	if err != nil {
		render500(w)
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

	var raw []rawPost

	for rows.Next() {
		var rp rawPost
		if err := rows.Scan(
			&rp.ID, &rp.UserID, &rp.Username,
			&rp.Title, &rp.Content, &rp.CreatedAt,
		); err == nil {
			raw = append(raw, rp)
		}
	}
	rows.Close()

	// Convert raw → full post model
	var postsList []models.Post

	for _, rp := range raw {
		p := models.Post{
			ID:        rp.ID,
			UserID:    rp.UserID,
			Username:  rp.Username,
			Title:     rp.Title,
			Content:   rp.Content,
			CreatedAt: rp.CreatedAt,
		}

		// Categories
		cats, _ := posts.GetCategoriesForPost(p.ID)
		for _, c := range cats {
			p.Categories = append(p.Categories, models.Category{
				ID:   c.ID,
				Name: c.Name,
			})
		}

		// Likes / Dislikes
		p.Likes, p.Dislikes = likes.CountPostLikes(p.ID)

		postsList = append(postsList, p)
	}

	// --------------------------------------------------------
	// LOAD CATEGORY LIST (for filters)
	// --------------------------------------------------------
	catRows, err := database.DB.Query(`
		SELECT id, name FROM categories ORDER BY name ASC
	`)
	if err != nil {
		render500(w)
		return
	}
	defer catRows.Close()

	var allCats []models.Category
	for catRows.Next() {
		var c models.Category
		if err := catRows.Scan(&c.ID, &c.Name); err == nil {
			allCats = append(allCats, c)
		}
	}

	// Data → Template
	data := HomeData{
		User:       user,
		Posts:      postsList,
		Categories: allCats,
	}

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		render500(w)
	}
}
