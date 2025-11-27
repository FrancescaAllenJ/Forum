package comments

import (
	"fmt"
	"net/http"
	"strconv"

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

	// Parse form values
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Check if user is logged in
	user, _ := auth.GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Read form values
	postIDStr := r.FormValue("post_id")
	content := r.FormValue("content")

	// Validate input
	if postIDStr == "" || content == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	// Convert post_id from string → int
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Insert comment into DB
	_, err = database.DB.Exec(`
		INSERT INTO comments (post_id, user_id, content)
		VALUES (?, ?, ?)
	`, postID, user.ID, content)
	if err != nil {
		// Internal DB error → log + panic → main.go wrapper → 500.html
		fmt.Println("Error inserting comment:", err)
		panic(err)
	}

	// Redirect back to the post page
	http.Redirect(w, r, "/post?id="+postIDStr, http.StatusSeeOther)
}
