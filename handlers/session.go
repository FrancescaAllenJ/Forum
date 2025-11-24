package auth

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"forum/database"
)

// SessionUser represents the logged-in user loaded from a session.
type SessionUser struct {
	ID       int
	Username string
	Email    string
}

// GetUserFromRequest checks the session cookie and loads the user if logged in.
func GetUserFromRequest(r *http.Request) (*SessionUser, error) {
	// Get cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// No cookie = not logged in
		return nil, nil
	}

	sessionID := cookie.Value
	if sessionID == "" {
		return nil, nil
	}

	// Look up session and user in DB
	var (
		userID    int
		username  string
		email     string
		expiresAt time.Time
	)

	err = database.DB.QueryRow(`
		SELECT users.id, users.username, users.email, sessions.expires_at
		FROM sessions
		JOIN users ON sessions.user_id = users.id
		WHERE sessions.id = ?
	`, sessionID).Scan(&userID, &username, &email, &expiresAt)

	if err == sql.ErrNoRows {
		// Session not found or user deleted
		return nil, nil
	}
	if err != nil {
		// Real DB error
		return nil, err
	}

	// Check if session expired
	if time.Now().After(expiresAt) {
		// Optionally: clean up expired session
		_, delErr := database.DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		if delErr != nil {
			log.Println("Error deleting expired session:", delErr)
		}
		return nil, nil
	}

	return &SessionUser{
		ID:       userID,
		Username: username,
		Email:    email,
	}, nil
}

// LogoutHandler destroys the session and clears the cookie.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get cookie (if no cookie, just redirect)
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		sessionID := cookie.Value

		// Delete from DB
		_, err := database.DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		if err != nil {
			log.Println("Error deleting session on logout:", err)
		}

		// Expire the cookie in the browser
		expiredCookie := http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
		}
		http.SetCookie(w, &expiredCookie)
	}

	// Redirect to home after logout
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
