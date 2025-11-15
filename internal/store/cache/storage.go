package cache

import (
	"context"

	"github.com/SAURABH200301/Social/internal/store"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64) (*store.Users, error)
		Set(context.Context, *store.Users) error
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{
			rdb: rdb,
		},
	}
}
