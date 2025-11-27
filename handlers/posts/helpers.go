package posts

import (
	"log"

	"forum/database"
)

// GetCategoriesForPost returns all categories for a given post ID.
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
