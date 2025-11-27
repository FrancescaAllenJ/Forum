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
	// Init DB + templates
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
	// GLOBAL WRAPPER
	// - custom 404
	// - panic → 500
	// --------------------------
	wrappedMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Recover from panics → 500
		defer func() {
			if rec := recover(); rec != nil {
				log.Println("Recovered panic:", rec)
				render500(w)
			}
		}()

		// Check if route exists
		_, pattern := mux.Handler(r)
		if pattern == "" {
			w.WriteHeader(http.StatusNotFound)
			templates.ExecuteTemplate(w, "error_404.html", nil)
			return
		}

		// Serve normally
		mux.ServeHTTP(w, r)
	})

	log.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", wrappedMux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// ----------------------------------------------------
// LOAD ALL TEMPLATES
// ----------------------------------------------------
func loadTemplates() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
}

// ----------------------------------------------------
// RENDER 500 PAGE
// ----------------------------------------------------
func render500(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	templates.ExecuteTemplate(w, "error_500.html", nil)
}

// ----------------------------------------------------
// HOMEPAGE HANDLER (IMPORT-CYCLE SAFE)
// ----------------------------------------------------
func homeHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.GetUserFromRequest(r)

	// STEP 1 — Load posts (raw)
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

	var rawPosts []rawPost

	for rows.Next() {
		var rp rawPost
		if err := rows.Scan(
			&rp.ID, &rp.UserID, &rp.Username, &rp.Title, &rp.Content, &rp.CreatedAt,
		); err != nil {
			continue
		}
		rawPosts = append(rawPosts, rp)
	}
	rows.Close()

	// STEP 2 — Convert raw posts → models.Post
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

		// Convert categories (handler → model)
		handlerCats, err := posts.GetCategoriesForPost(p.ID)
		if err == nil {
			var converted []models.Category
			for _, c := range handlerCats {
				converted = append(converted, models.Category{
					ID:   c.ID,
					Name: c.Name,
				})
			}
			p.Categories = converted
		}

		// Load like/dislike counts
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
		render500(w)
	}
}
