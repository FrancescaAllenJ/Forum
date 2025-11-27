package posts

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"forum/database"
	auth "forum/handlers"
	likes "forum/handlers/likes"
	"forum/models"
)

var postTmpl = template.Must(template.ParseGlob("templates/*.html"))

type PostPageData struct {
	User     *auth.SessionUser
	Post     models.Post
	Comments []CommentView
}

type CommentView struct {
	ID        int
	Content   string
	CreatedAt string
	Username  string
	Likes     int
	Dislikes  int
}

// -----------------------------------------------------------
// ViewPostHandler — loads a single post with comments + likes
// -----------------------------------------------------------
func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.GetUserFromRequest(r)

	// 1. Get post ID from query string
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// 2. Load the post
	var p models.Post
	err = database.DB.QueryRow(`
		SELECT posts.id,
		       posts.user_id,
		       users.username,
		       posts.title,
		       posts.content,
		       strftime('%Y-%m-%d %H:%M:%S', posts.created_at)
		FROM posts
		JOIN users ON posts.user_id = users.id
		WHERE posts.id = ?
	`, postID).Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.CreatedAt)

	if err != nil {
		// Treat as "not found" for now (could be sql.ErrNoRows or other)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// 3. Load & convert categories (handler type → models.Category)
	handlerCats, _ := GetCategoriesForPost(postID)
	var convertedCats []models.Category
	for _, c := range handlerCats {
		convertedCats = append(convertedCats, models.Category{
			ID:   c.ID,
			Name: c.Name,
		})
	}
	p.Categories = convertedCats

	// 4. Likes/dislikes for the post
	lc, dc := likes.CountPostLikes(postID)
	p.Likes = lc
	p.Dislikes = dc

	// 5. Load comments into a buffer (avoid SQLite cursor + nested queries)
	rows, err := database.DB.Query(`
		SELECT comments.id,
		       comments.content,
		       strftime('%Y-%m-%d %H:%M:%S', comments.created_at),
		       users.username
		FROM comments
		JOIN users ON comments.user_id = users.id
		WHERE comments.post_id = ?
		ORDER BY comments.created_at ASC
	`, postID)
	if err != nil {
		// Internal DB failure → panic, caught by main.go wrapper → 500 page
		panic(err)
	}

	type rawComment struct {
		ID        int
		Content   string
		CreatedAt string
		Username  string
	}

	var rawComments []rawComment
	for rows.Next() {
		var rc rawComment
		if err := rows.Scan(&rc.ID, &rc.Content, &rc.CreatedAt, &rc.Username); err != nil {
			log.Println("Error scanning comment:", err)
			continue
		}
		rawComments = append(rawComments, rc)
	}
	rows.Close()

	// 6. Convert to CommentView + per-comment likes
	var comments []CommentView
	for _, rc := range rawComments {
		cv := CommentView{
			ID:        rc.ID,
			Content:   rc.Content,
			CreatedAt: rc.CreatedAt,
			Username:  rc.Username,
		}

		lc, dc := likes.CountCommentLikes(rc.ID)
		cv.Likes = lc
		cv.Dislikes = dc

		comments = append(comments, cv)
	}

	// 7. Render template
	data := PostPageData{
		User:     user,
		Post:     p,
		Comments: comments,
	}

	if err := postTmpl.ExecuteTemplate(w, "post.html", data); err != nil {
		// Template failure → panic → main.go wrapper → 500.html
		panic(err)
	}
}
