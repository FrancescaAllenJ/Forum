package posts

import (
	"html/template"
	"log"
	"net/http"

	"forum/database"
	auth "forum/handlers"
)

var postTmpl = template.Must(template.ParseGlob("templates/*.html"))

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

	// ---------------------------------------------------------
	// GET METHOD — show the create post form with categories
	// ---------------------------------------------------------
	case "GET":

		// Load all categories for selection
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

		// Pass categories to template
		data := map[string]interface{}{
			"Categories": cats,
		}

		postTmpl.ExecuteTemplate(w, "create_post.html", data)

	// ---------------------------------------------------------
	// POST METHOD — create a post and save its categories
	// ---------------------------------------------------------
	case "POST":

		title := r.FormValue("title")
		content := r.FormValue("content")

		// Read multiple category IDs (checkboxes)
		categoryIDs := r.Form["category_ids"]

		if title == "" || content == "" {
			postTmpl.ExecuteTemplate(w, "create_post.html",
				map[string]string{"Error": "All fields are required."})
			return
		}

		// Insert new post
		result, err := database.DB.Exec(`
			INSERT INTO posts (user_id, title, content)
			VALUES (?, ?, ?)
		`, user.ID, title, content)

		if err != nil {
			log.Println("Insert post error:", err)
			http.Error(w, "Error saving post", http.StatusInternalServerError)
			return
		}

		// Get ID of new post
		postID, _ := result.LastInsertId()

		// Insert post-category relationships
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
