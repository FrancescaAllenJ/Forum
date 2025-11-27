package categories

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"forum/database"
	auth "forum/handlers"
	posts "forum/handlers/posts"
	"forum/models"
)

var categoryTmpl = template.Must(template.ParseGlob("templates/*.html"))

type CategoryPageData struct {
	User     *auth.SessionUser
	Category models.Category
	Posts    []models.Post
}

func ViewCategoryHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.GetUserFromRequest(r)

	// ---------------------------------------------------------
	// 1. Get the category ID from URL
	// ---------------------------------------------------------
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing category ID", http.StatusBadRequest)
		return
	}

	catID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// ---------------------------------------------------------
	// 2. Load category info (convert to models.Category)
	// ---------------------------------------------------------
	var cat models.Category
	err = database.DB.QueryRow(`
		SELECT id, name
		FROM categories
		WHERE id = ?
	`, catID).Scan(&cat.ID, &cat.Name)

	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	// ---------------------------------------------------------
	// 3. Load all posts inside this category
	// ---------------------------------------------------------
	rows, err := database.DB.Query(`
		SELECT posts.id,
		       posts.user_id,
		       users.username,
		       posts.title,
		       posts.content,
		       strftime('%Y-%m-%d %H:%M:%S', posts.created_at)
		FROM posts
		JOIN users ON posts.user_id = users.id
		JOIN post_categories ON post_categories.post_id = posts.id
		WHERE post_categories.category_id = ?
		ORDER BY posts.created_at DESC
	`, catID)

	if err != nil {
		http.Error(w, "Error loading posts", http.StatusInternalServerError)
		return
	}

	var rawPosts []struct {
		ID        int
		UserID    int
		Username  string
		Title     string
		Content   string
		CreatedAt string
	}

	for rows.Next() {
		var rp struct {
			ID        int
			UserID    int
			Username  string
			Title     string
			Content   string
			CreatedAt string
		}

		if err := rows.Scan(&rp.ID, &rp.UserID, &rp.Username, &rp.Title, &rp.Content, &rp.CreatedAt); err != nil {
			log.Println("SCAN ERROR:", err)
			continue
		}

		rawPosts = append(rawPosts, rp)
	}

	rows.Close() // IMPORTANT for SQLite

	// ---------------------------------------------------------
	// 4. Convert raw posts → []models.Post (with categories)
	// ---------------------------------------------------------
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

		// Load categories for each post (convert to models.Category)
		cats, _ := posts.GetCategoriesForPost(p.ID)

		var convertedCats []models.Category
		for _, c := range cats {
			convertedCats = append(convertedCats, models.Category{
				ID:   c.ID,
				Name: c.Name,
			})
		}
		p.Categories = convertedCats

		postsList = append(postsList, p)
	}

	// ---------------------------------------------------------
	// 5. Render template
	// ---------------------------------------------------------
	data := CategoryPageData{
		User:     user,
		Category: cat,
		Posts:    postsList,
	}

	if err := categoryTmpl.ExecuteTemplate(w, "category.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}
