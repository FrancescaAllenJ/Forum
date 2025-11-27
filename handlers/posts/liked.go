package posts

import (
	"html/template"
	"net/http"

	"forum/database"
	auth "forum/handlers"
	likes "forum/handlers/likes"
	"forum/models"
)

var likedPostsTmpl = template.Must(template.ParseGlob("templates/*.html"))

// LikedPostsHandler shows posts that the current user has liked (value = 1).
func LikedPostsHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 1. Load raw posts liked by this user
	rows, err := database.DB.Query(`
		SELECT posts.id,
		       posts.user_id,
		       users.username,
		       posts.title,
		       posts.content,
		       strftime('%Y-%m-%d %H:%M:%S', posts.created_at)
		FROM posts
		JOIN likes ON likes.post_id = posts.id
		JOIN users ON posts.user_id = users.id
		WHERE likes.user_id = ? AND likes.value = 1
		ORDER BY posts.created_at DESC
	`, user.ID)
	if err != nil {
		// Internal DB error → panic → 500.html
		panic(err)
	}

	type rawPost struct {
		ID        int
		UserID    int
		Username  string
		Title     string
		Content   string
		CreatedAt string
	}

	var rawPosts []rawPost
	for rows.Next() {
		var rp rawPost
		if err := rows.Scan(&rp.ID, &rp.UserID, &rp.Username, &rp.Title, &rp.Content, &rp.CreatedAt); err != nil {
			continue
		}
		rawPosts = append(rawPosts, rp)
	}
	rows.Close()

	// 2. Enrich with categories + likes
	var postsList []models.Post

	for _, rp := range rawPosts {
		p := models.Post{
			ID:        rp.ID,
			UserID:    rp.UserID,
			Username:  rp.Username,
			Title:     rp.Title,
			Content:   rp.Content,
			CreatedAt: rp.CreatedAt,
		}

		// Categories
		handlerCats, _ := GetCategoriesForPost(p.ID)
		var converted []models.Category
		for _, c := range handlerCats {
			converted = append(converted, models.Category{
				ID:   c.ID,
				Name: c.Name,
			})
		}
		p.Categories = converted

		// Likes
		l, d := likes.CountPostLikes(p.ID)
		p.Likes = l
		p.Dislikes = d

		postsList = append(postsList, p)
	}

	// 3. Render template
	data := struct {
		User  *auth.SessionUser
		Posts []models.Post
	}{
		User:  user,
		Posts: postsList,
	}

	if err := likedPostsTmpl.ExecuteTemplate(w, "liked_posts.html", data); err != nil {
		// Template error → panic → 500.html
		panic(err)
	}
}
