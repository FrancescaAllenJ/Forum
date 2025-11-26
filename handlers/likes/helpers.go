package likes

import "forum/database"

func CountLikes(postID int, commentID int) (int, int) {
	var likes, dislikes int

	// COUNT LIKES
	database.DB.QueryRow(`
        SELECT COUNT(*) FROM likes 
        WHERE value = 1 AND post_id = ? AND comment_id IS ?
    `, postID, commentID).Scan(&likes)

	// COUNT DISLIKES
	database.DB.QueryRow(`
        SELECT COUNT(*) FROM likes 
        WHERE value = -1 AND post_id = ? AND comment_id IS ?
    `, postID, commentID).Scan(&dislikes)

	return likes, dislikes
}
