package likes

import "forum/database"

// CountPostLikes returns (likes, dislikes) for a given post.
func CountPostLikes(postID int) (int, int) {
	var likesCount, dislikesCount int

	// Count likes for this post
	database.DB.QueryRow(`
        SELECT COUNT(*) FROM likes
        WHERE value = 1 AND post_id = ? AND comment_id IS NULL
    `, postID).Scan(&likesCount)

	// Count dislikes for this post
	database.DB.QueryRow(`
        SELECT COUNT(*) FROM likes
        WHERE value = -1 AND post_id = ? AND comment_id IS NULL
    `, postID).Scan(&dislikesCount)

	return likesCount, dislikesCount
}

// CountCommentLikes returns (likes, dislikes) for a given comment.
func CountCommentLikes(commentID int) (int, int) {
	var likesCount, dislikesCount int

	// Count likes for this comment
	database.DB.QueryRow(`
        SELECT COUNT(*) FROM likes
        WHERE value = 1 AND comment_id = ? AND post_id IS NULL
    `, commentID).Scan(&likesCount)

	// Count dislikes for this comment
	database.DB.QueryRow(`
        SELECT COUNT(*) FROM likes
        WHERE value = -1 AND comment_id = ? AND post_id IS NULL
    `, commentID).Scan(&dislikesCount)

	return likesCount, dislikesCount
}
