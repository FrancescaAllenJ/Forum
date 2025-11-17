package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func main() {
	http.HandleFunc("/", homeHandler)

	// Serve static files like CSS
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

// WHAT A WEB SERVER ACTUALLY IS

// A web server is a program that:

// Listens for incoming network requests

// Understands HTTP

// Decides which function to run for each URL

// Sends responses back to the browser

// Your Go code is doing all four.

// net/http - STAR OF THE SHOW
// create a server, listen on a port, read requests, send responses, handle cookies, manage HTTP methods (GET, POST)
// define handlers, This is why Go is famous for web development — the HTTP server is built into the language.
