package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SAURABH200301/Social/internal/store"
	"github.com/redis/go-redis/v9"
)

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute

func (s *UserStore) Get(ctx context.Context, userID int64) (*store.Users, error) {
	cacheKey := fmt.Sprintf("user-%v", userID)
	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}
	var user store.Users
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.Users) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)
	json, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.rdb.SetEx(ctx, cacheKey, json, UserExpTime).Err()
}
