package posts

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"forum/database"
	auth "forum/handlers"
	likes "forum/handlers/likes"
)

var postTmpl = template.Must(template.ParseGlob("templates/*.html"))

type PostPageData struct {
	User     *auth.SessionUser
	Post     SinglePost
	Comments []CommentView
}

type SinglePost struct {
	ID         int
	Title      string
	Content    string
	CreatedAt  string
	Username   string
	Likes      int
	Dislikes   int
	Categories []Category
}

type CommentView struct {
	ID        int
	Content   string
	CreatedAt string
	Username  string
	Likes     int
	Dislikes  int
}

// ViewPostHandler shows a single post with comments, categories, and likes.
func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.GetUserFromRequest(r)

	// ---------------------------------------------------------
	// 1. Read post ID
	// ---------------------------------------------------------
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

	// ---------------------------------------------------------
	// 2. Load the post (use strftime for safe Go string scan)
	// ---------------------------------------------------------
	var post SinglePost
	err = database.DB.QueryRow(`
		SELECT posts.id,
		       posts.title,
		       posts.content,
		       strftime('%Y-%m-%d %H:%M:%S', posts.created_at) AS created_at,
		       users.username
		FROM posts
		JOIN users ON posts.user_id = users.id
		WHERE posts.id = ?
	`, postID).Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.Username)

	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// ---------------------------------------------------------
	// 3. Load categories for this post
	// ---------------------------------------------------------
	cats, _ := GetCategoriesForPost(postID)
	post.Categories = cats

	// ---------------------------------------------------------
	// 4. Load likes/dislikes for this post
	// ---------------------------------------------------------
	postLikes, postDislikes := likes.CountPostLikes(postID)
	post.Likes = postLikes
	post.Dislikes = postDislikes

	// ---------------------------------------------------------
	// 5. Load comments (BUFFER FIRST to avoid SQLite deadlock)
	// ---------------------------------------------------------
	rows, err := database.DB.Query(`
		SELECT comments.id,
		       comments.content,
		       strftime('%Y-%m-%d %H:%M:%S', comments.created_at) AS created_at,
		       users.username
		FROM comments
		JOIN users ON comments.user_id = users.id
		WHERE comments.post_id = ?
		ORDER BY comments.created_at ASC
	`, postID)

	if err != nil {
		http.Error(w, "Error loading comments", http.StatusInternalServerError)
		return
	}

	// Temporary buffer structure
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

	rows.Close() // IMPORTANT: close before nested DB queries

	// ---------------------------------------------------------
	// 6. Build final CommentView list with likes
	// ---------------------------------------------------------
	var comments []CommentView

	for _, rc := range rawComments {
		c := CommentView{
			ID:        rc.ID,
			Content:   rc.Content,
			CreatedAt: rc.CreatedAt,
			Username:  rc.Username,
		}

		cl, cd := likes.CountCommentLikes(c.ID)
		c.Likes = cl
		c.Dislikes = cd

		comments = append(comments, c)
	}

	// ---------------------------------------------------------
	// 7. Render template
	// ---------------------------------------------------------
	data := PostPageData{
		User:     user,
		Post:     post,
		Comments: comments,
	}

	if err := postTmpl.ExecuteTemplate(w, "post.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}
