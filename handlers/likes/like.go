package likes

import (
	"net/http"
	"strconv"

	"forum/database"
	auth "forum/handlers"
)

// LikeHandler handles likes & dislikes for posts and comments.
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Check login
	user, _ := auth.GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	targetType := r.FormValue("type") // "post" or "comment"
	targetID := r.FormValue("id")
	value := r.FormValue("value")

	val, err := strconv.Atoi(value)
	if err != nil || (val != 1 && val != -1) {
		http.Error(w, "Invalid like value", http.StatusBadRequest)
		return
	}

	switch targetType {
	case "post":
		handlePostLike(w, r, user.ID, targetID, val)
	case "comment":
		handleCommentLike(w, r, user.ID, targetID, val)
	default:
		http.Error(w, "Invalid like target", http.StatusBadRequest)
	}
}

// -------------------------------------------------------
// POST LIKE
// -------------------------------------------------------
func handlePostLike(w http.ResponseWriter, r *http.Request, userID int, targetID string, val int) {
	postID, err := strconv.Atoi(targetID)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var existing int
	err = database.DB.QueryRow(`
        SELECT value FROM likes
        WHERE user_id = ? AND post_id = ? AND comment_id IS NULL
    `, userID, postID).Scan(&existing)

	if err == nil && existing == val {
		database.DB.Exec(`
            DELETE FROM likes
            WHERE user_id = ? AND post_id = ? AND comment_id IS NULL
        `, userID, postID)
	} else if err == nil {
		database.DB.Exec(`
            UPDATE likes SET value = ?
            WHERE user_id = ? AND post_id = ? AND comment_id IS NULL
        `, val, userID, postID)
	} else {
		database.DB.Exec(`
            INSERT INTO likes (user_id, post_id, comment_id, value)
            VALUES (?, ?, NULL, ?)
        `, userID, postID, val)
	}

	// Always redirect safely to post page (fixes Back button)
	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

// -------------------------------------------------------
// COMMENT LIKE
// -------------------------------------------------------
func handleCommentLike(w http.ResponseWriter, r *http.Request, userID int, targetID string, val int) {
	commentID, err := strconv.Atoi(targetID)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	var existing int
	err = database.DB.QueryRow(`
        SELECT value FROM likes
        WHERE user_id = ? AND comment_id = ? AND post_id IS NULL
    `, userID, commentID).Scan(&existing)

	if err == nil && existing == val {
		database.DB.Exec(`
            DELETE FROM likes
            WHERE user_id = ? AND comment_id = ? AND post_id IS NULL
        `, userID, commentID)
	} else if err == nil {
		database.DB.Exec(`
            UPDATE likes SET value = ?
            WHERE user_id = ? AND comment_id = ? AND post_id IS NULL
        `, val, userID, commentID)
	} else {
		database.DB.Exec(`
            INSERT INTO likes (user_id, post_id, comment_id, value)
            VALUES (?, NULL, ?, ?)
        `, userID, commentID, val)
	}

	// We need the post ID → extract from referer:
	ref := r.Header.Get("Referer")

	// Fallback → redirect home if empty
	if ref == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Safe redirect back to the post page URL
	http.Redirect(w, r, ref, http.StatusSeeOther)
}
