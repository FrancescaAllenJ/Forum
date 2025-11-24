package auth

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"forum/database"

	"golang.org/x/crypto/bcrypt"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

// RegisterHandler handles GET + POST for user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tmpl.ExecuteTemplate(w, "register.html", nil)
	case "POST":
		handleRegisterPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleRegisterPost(w http.ResponseWriter, r *http.Request) {
	// Parse form values
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	// BASIC VALIDATION
	if email == "" || username == "" || password == "" {
		tmpl.ExecuteTemplate(w, "register.html", "All fields are required.")
		return
	}

	// CHECK EMAIL EXISTS
	var exists int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error checking email:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if exists > 0 {
		tmpl.ExecuteTemplate(w, "register.html", "Email is already registered.")
		return
	}

	// CHECK USERNAME EXISTS
	err = database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error checking username:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if exists > 0 {
		tmpl.ExecuteTemplate(w, "register.html", "Username is already taken.")
		return
	}

	// HASH PASSWORD
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// INSERT USER
	_, err = database.DB.Exec(
		"INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
		email, username, hashed,
	)

	if err != nil {
		log.Println("Insert user error:", err)
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	// REDIRECT TO LOGIN
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
