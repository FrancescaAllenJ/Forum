package posts

import (
	"log"

	"forum/database"
)

// Category represents a post category.
type Category struct {
	ID   int
	Name string
}

// GetCategoriesForPost returns all categories linked to a post.
func GetCategoriesForPost(postID int) ([]Category, error) {
	rows, err := database.DB.Query(`
        SELECT categories.id, categories.name
        FROM categories
        JOIN post_categories ON categories.id = post_categories.category_id
        WHERE post_categories.post_id = ?
    `, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Category

	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			log.Println("Error scanning category:", err)
			continue
		}
		result = append(result, c)
	}

	return result, nil
}
