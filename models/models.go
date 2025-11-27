package models

// Post represents a forum post displayed across pages.
type Post struct {
	ID         int
	UserID     int
	Username   string
	Title      string
	Content    string
	CreatedAt  string
	Likes      int
	Dislikes   int
	Categories []Category
}

// Comment represents a single comment under a post.
type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Username  string
	Content   string
	CreatedAt string
	Likes     int
	Dislikes  int
}

// Category represents a single category.
// This MUST NOT import handlers or cause cycles.
type Category struct {
	ID   int
	Name string
}
