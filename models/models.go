package models

import posts "forum/handlers/posts"

// Post represents a forum post displayed on the homepage or post page.
type Post struct {
	ID         int
	UserID     int
	Username   string
	Title      string
	Content    string
	CreatedAt  string
	Categories []posts.Category // categories linked to the post
}

// Comment represents a single comment under a post.
type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Username  string
	Content   string
	CreatedAt string
}

// Category is a lightweight passthrough of the posts.Category struct.
// We embed it so templates can use either type without conflict.
type Category = posts.Category
