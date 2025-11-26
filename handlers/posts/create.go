package posts

import (
	"log"
	"net/http"

	"forum/database"
	auth "forum/handlers"
)

// For passing category list to the template
type Category struct {
	ID   int
	Name string
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	// Check login
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	switch r.Method {

	// -------------------------
	// GET — Show form
	// -------------------------
	case "GET":

		rows, err := database.DB.Query(`SELECT id, name FROM categories`)
		if err != nil {
			http.Error(w, "Error loading categories", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var cats []Category
		for rows.Next() {
			var c Category
			rows.Scan(&c.ID, &c.Name)
			cats = append(cats, c)
		}

		data := map[string]interface{}{
			"Categories": cats,
		}

		postTmpl.ExecuteTemplate(w, "create_post.html", data)

	// -------------------------
	// POST — Create post
	// -------------------------
	case "POST":

		title := r.FormValue("title")
		content := r.FormValue("content")
		categoryIDs := r.Form["category_ids"]

		if title == "" || content == "" {
			postTmpl.ExecuteTemplate(w, "create_post.html",
				map[string]string{"Error": "All fields are required."})
			return
		}

		result, err := database.DB.Exec(`
			INSERT INTO posts (user_id, title, content)
			VALUES (?, ?, ?)
		`, user.ID, title, content)

		if err != nil {
			log.Println("Insert post error:", err)
			http.Error(w, "Error saving post", http.StatusInternalServerError)
			return
		}

		postID, _ := result.LastInsertId()

		for _, catID := range categoryIDs {
			_, err := database.DB.Exec(`
				INSERT INTO post_categories (post_id, category_id)
				VALUES (?, ?)
			`, postID, catID)

			if err != nil {
				log.Println("Error inserting post-category:", err)
			}
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
