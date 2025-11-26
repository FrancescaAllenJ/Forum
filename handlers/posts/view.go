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
	Categories []Category // <-- ✅ ADDED
}

type CommentView struct {
	ID        int
	Content   string
	CreatedAt string
	Username  string
	Likes     int
	Dislikes  int
}

func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.GetUserFromRequest(r)

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

	var post SinglePost
	err = database.DB.QueryRow(`
		SELECT posts.id, posts.title, posts.content, posts.created_at, users.username
		FROM posts
		JOIN users ON posts.user_id = users.id
		WHERE posts.id = ?
	`, postID).Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.Username)

	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// --------------------------------------------
	// ✅ NEW — Load categories for this post
	// --------------------------------------------
	cats, _ := GetCategoriesForPost(postID)
	post.Categories = cats

	// Load likes/dislikes for the post
	postLikes, postDislikes := likes.CountPostLikes(postID)
	post.Likes = postLikes
	post.Dislikes = postDislikes

	// Load comments
	rows, err := database.DB.Query(`
		SELECT comments.id, comments.content, comments.created_at, users.username
		FROM comments
		JOIN users ON comments.user_id = users.id
		WHERE comments.post_id = ?
		ORDER BY comments.created_at ASC
	`, postID)

	if err != nil {
		http.Error(w, "Error loading comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []CommentView
	for rows.Next() {
		var c CommentView
		if err := rows.Scan(&c.ID, &c.Content, &c.CreatedAt, &c.Username); err != nil {
			log.Println("Error scanning comment:", err)
			continue
		}

		// Load likes/dislikes for each comment
		cl, cd := likes.CountCommentLikes(c.ID)
		c.Likes = cl
		c.Dislikes = cd

		comments = append(comments, c)
	}

	data := PostPageData{
		User:     user,
		Post:     post,
		Comments: comments,
	}

	if err := postTmpl.ExecuteTemplate(w, "post.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
