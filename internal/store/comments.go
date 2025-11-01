package store

import (
	"context"
	"database/sql"
	"fmt"
)

type Comments struct {
	ID        int32  `json:"id"`
	PostID    int32  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      Users  `json:"user"`
}
type CommentsStore struct {
	db *sql.DB
}

func (s *CommentsStore) GetCommentsByPostID(ctx context.Context, postID int32) ([]*Comments, error) {
	query := `SELECT comments.id, comments.content, comments.post_id, comments.user_id, comments.created_at, users.username FROM comments JOIN users ON users.id = comments.user_id WHERE post_id = $1 ORDER BY comments.created_at DESC`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("record not found")
		}
		return nil, err
	}
	defer rows.Close()

	var comments []*Comments
	for rows.Next() {
		var comment Comments
		if err := rows.Scan(&comment.ID, &comment.Content, &comment.PostID, &comment.UserID, &comment.CreatedAt, &comment.User.Username); err != nil {
			return nil, err
		}
		comment.User.ID = comment.UserID
		comments = append(comments, &comment)
	}
	return comments, nil
}
