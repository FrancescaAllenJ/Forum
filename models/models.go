package models

import posts "forum/handlers/posts"

// Post represents a forum post displayed on the homepage or post page.
// CreatedAt is kept as a string because your SQL query will return it
// as a formatted TEXT value using strftime(...).
type Post struct {
	ID         int
	UserID     int
	Username   string
	Title      string
	Content    string
	CreatedAt  string // string is OK (SQL query returns TEXT)
	Likes      int
	Dislikes   int
	Categories []posts.Category
}

// Comment represents a single comment under a post.
type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Username  string
	Content   string
	CreatedAt string // also fine as string
}

// Category is a passthrough alias to posts.Category.
type Category = posts.Category
