package likes

import (
	"net/http"
	"strconv"

	"forum/database"
	auth "forum/handlers"
)

// LikeHandler handles both likes and dislikes for posts and comments.
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract form data
	targetType := r.FormValue("type") // "post" or "comment"
	targetID := r.FormValue("id")     // post_id or comment_id
	value := r.FormValue("value")     // "1" (like) or "-1" (dislike)

	val, err := strconv.Atoi(value)
	if err != nil || (val != 1 && val != -1) {
		http.Error(w, "Invalid like value", http.StatusBadRequest)
		return
	}

	// Determine if target is post or comment
	var postID, commentID interface{}

	if targetType == "post" {
		postID, err = strconv.Atoi(targetID)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		commentID = nil
	} else if targetType == "comment" {
		commentID, err = strconv.Atoi(targetID)
		if err != nil {
			http.Error(w, "Invalid comment ID", http.StatusBadRequest)
			return
		}
		postID = nil
	} else {
		http.Error(w, "Invalid like target", http.StatusBadRequest)
		return
	}

	// Check if user already liked/disliked this item
	var existingValue int
	err = database.DB.QueryRow(`
        SELECT value FROM likes 
        WHERE user_id = ? AND post_id IS ? AND comment_id IS ?
    `, user.ID, postID, commentID).Scan(&existingValue)

	// CASE 1: User already liked the same thing → remove it (toggle off)
	if err == nil && existingValue == val {
		database.DB.Exec(`
            DELETE FROM likes 
            WHERE user_id = ? AND post_id IS ? AND comment_id IS ?
        `, user.ID, postID, commentID)

	} else if err == nil {
		// CASE 2: User switches from like → dislike or dislike → like
		database.DB.Exec(`
            UPDATE likes SET value = ?
            WHERE user_id = ? AND post_id IS ? AND comment_id IS ?
        `, val, user.ID, postID, commentID)

	} else {
		// CASE 3: No existing like/dislike → insert new
		database.DB.Exec(`
            INSERT INTO likes (user_id, post_id, comment_id, value)
            VALUES (?, ?, ?, ?)
        `, user.ID, postID, commentID, val)
	}

	// Redirect back to previous page
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
