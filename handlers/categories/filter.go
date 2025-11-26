package categories

import (
	"html/template"
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
	Category posts.Category
	Posts    []models.Post
}

func ViewCategoryHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.GetUserFromRequest(r)

	// 1. Get the category ID from URL
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

	// 2. Load category info (name)
	var category posts.Category
	err = database.DB.QueryRow(`
        SELECT id, name
        FROM categories
        WHERE id = ?
    `, catID).Scan(&category.ID, &category.Name)

	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	// 3. Load all posts inside this category
	rows, err := database.DB.Query(`
        SELECT posts.id, posts.user_id, users.username, posts.title, posts.content, posts.created_at
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
	defer rows.Close()

	var postsList []models.Post

	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.CreatedAt)
		if err != nil {
			continue
		}

		// Load categories for each post
		cats, _ := posts.GetCategoriesForPost(p.ID)
		p.Categories = cats

		postsList = append(postsList, p)
	}

	// 4. Build data to pass to template
	data := CategoryPageData{
		User:     user,
		Category: category,
		Posts:    postsList,
	}

	// 5. Render template
	err = categoryTmpl.ExecuteTemplate(w, "category.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
