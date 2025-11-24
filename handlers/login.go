package auth

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"forum/database"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var loginTmpl = template.Must(template.ParseGlob("templates/*.html"))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		loginTmpl.ExecuteTemplate(w, "login.html", nil)
	case "POST":
		handleLoginPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleLoginPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		loginTmpl.ExecuteTemplate(w, "login.html", "All fields are required.")
		return
	}

	// Check if user exists
	var (
		userID   int
		username string
		hash     string
	)

	err := database.DB.QueryRow(
		"SELECT id, username, password FROM users WHERE email = ?", email,
	).Scan(&userID, &username, &hash)

	if err == sql.ErrNoRows {
		loginTmpl.ExecuteTemplate(w, "login.html", "Invalid email or password.")
		return
	}

	if err != nil {
		log.Println("Login query error:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Compare bcrypt-hashed password
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		loginTmpl.ExecuteTemplate(w, "login.html", "Invalid email or password.")
		return
	}

	// Create session ID
	sessionID := uuid.New().String()
	expires := time.Now().Add(24 * time.Hour)

	// Insert into sessions table
	_, err = database.DB.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expires,
	)

	if err != nil {
		log.Println("Error creating session:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Set cookie
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expires,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // change to true for HTTPS
	}
	http.SetCookie(w, &cookie)

	// Redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
