package comments

import (
	"fmt"
	"net/http"

	"forum/database"
	auth "forum/handlers"
)

// CreateCommentHandler handles POST /create-comment
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check login
	user, _ := auth.GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Read form values
	postID := r.FormValue("post_id")
	content := r.FormValue("content")

	// Validate
	if postID == "" || content == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	// Insert comment into DB
	_, err := database.DB.Exec(`
		INSERT INTO comments (post_id, user_id, content)
		VALUES (?, ?, ?)
	`, postID, user.ID, content)

	if err != nil {
		fmt.Println("Error inserting comment:", err)
		http.Error(w, "Could not save comment", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post page
	http.Redirect(w, r, "/post?id="+postID, http.StatusSeeOther)
}
