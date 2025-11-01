package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type Post struct {
	ID        int32      `json:"id"`
	Content   string     `json:"content"`
	Title     string     `json:"title"`
	UserID    int64      `json:"user_id"`
	CreatedAt string     `json:"created_at"`
	Version   int32      `json:"version"`
	Tags      []string   `json:"tags,omitempty"`
	UserName  string     `json:"username,omitempty"`
	Comments  []Comments `json:"comments,omitempty"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (content, title, user_id, created_at, tags) 
			VALUES ($1, $2, $3, NOW(), $4) RETURNING id, created_at`
	var id int32

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, post.Content, post.Title, post.UserID, pq.Array(post.Tags)).Scan(&id, &post.CreatedAt)
	post.ID = id
	return err
}

func (s *PostStore) GetByID(ctx context.Context, id int32) (*Post, error) {
	query := `SELECT id, content, title, user_id, created_at, tags, version FROM posts WHERE id = $1`
	row := s.db.QueryRowContext(ctx, query, id)

	var post Post
	err := row.Scan(&post.ID, &post.Content, &post.Title, &post.UserID, &post.CreatedAt, pq.Array(&post.Tags), &post.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post with ID %d not found", id)
		}
		return nil, err
	}
	return &post, nil
}

func (s *PostStore) DeletePostByID(ctx context.Context, id int32) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("post with ID %d not found", id)
	}
	return nil
}

func (s *PostStore) UpdatePost(ctx context.Context, post *Post) error {
	query := `UPDATE posts SET content = $1, title = $2, tags = $3, version = version + 1 WHERE id = $4 AND version = $5 RETURNING version`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, post.Content, post.Title, pq.Array(post.Tags), post.ID, post.Version).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return fmt.Errorf("edit conflict for post with ID %d", post.ID)
		default:
			return err
		}
	}
	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userID int, pfq *PaginationFeedQuery) ([]PostWithMetadata, error) {
	query := `
		SELECT p.id, p.content, p.title, p.user_id, p.created_at, p.tags, p.version, u.username,
		COUNT(c.id) AS comments_count
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON u.id = p.user_id
		WHERE p.user_id = $1 AND (p.title ILIKE '%' || $5 || '%' OR p.content ILIKE '%' || $5 || '%') AND
		(p.tags @> $6 OR $6 = '{}')
		GROUP BY p.id, u.username
		ORDER BY p.created_at $2
		LIMIT $3 OFFSET $4
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, query, userID, pfq.Sort, pfq.Limit, pfq.Offset, pfq.Search, pq.Array(pfq.Tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feeds []PostWithMetadata
	for rows.Next() {
		var feed PostWithMetadata
		if err := rows.Scan(&feed.ID, &feed.Content, &feed.Title, &feed.UserID, &feed.CreatedAt, pq.Array(&feed.Tags), &feed.Version, &feed.UserName, &feed.CommentsCount); err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, nil
}
