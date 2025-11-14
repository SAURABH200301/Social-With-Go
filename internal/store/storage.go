package store

import (
	"context"
	"database/sql"
	"time"
)

var (
	QueryTimeOutDuration = 5 * time.Second
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int32) (*Post, error)
		DeletePostByID(context.Context, int32) error
		UpdatePost(context.Context, *Post) error
		GetUserFeed(context.Context, int, *PaginationFeedQuery) ([]PostWithMetadata, error)
	}
	Users interface {
		Create(context.Context, *sql.Tx, *Users) error
		GetByID(context.Context, int64) (*Users, error)
		CreateAndInvite(ctx context.Context, user *Users, token string, invitationExp time.Duration) error
		Activate(ctx context.Context, token string) error
		DeleteByID(ctx context.Context, id int64) error
		GetByEmail(ctx context.Context, email string) (*Users, error)
	}
	Comments interface {
		GetCommentsByPostID(context.Context, int32) ([]*Comments, error)
	}
	Roles interface {
		GetByName(context.Context, string) (*Role, error)
	}
}

func NewPostgresStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostStore{db: db},
		Users:    &UsersStorage{db: db},
		Comments: &CommentsStore{db: db},
		Roles:    &RoleStore{db: db},
	}
}

func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
