package posts

import (
	"html/template"
	"net/http"

	"forum/database"
	auth "forum/handlers"
)

var postTmpl = template.Must(template.ParseGlob("templates/*.html"))

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	// Check login state
	user, err := auth.GetUserFromRequest(r)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// Not logged in → redirect
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case "GET":
		postTmpl.ExecuteTemplate(w, "create_post.html", nil)

	case "POST":
		title := r.FormValue("title")
		content := r.FormValue("content")

		if title == "" || content == "" {
			postTmpl.ExecuteTemplate(w, "create_post.html",
				map[string]string{"Error": "All fields are required."})
			return
		}

		_, err := database.DB.Exec(
			"INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)",
			user.ID, title, content,
		)

		if err != nil {
			http.Error(w, "Error saving post", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
