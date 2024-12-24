package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ekachaikeaw/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute

func (s *UserStore) Get(c context.Context, userID int64) (*store.User, error) {
	key := fmt.Sprintf("user-%d", userID)

	data, err := s.rdb.Get(c, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)	
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) Set(c context.Context, user *store.User) error {
	key := fmt.Sprintf("user-%d", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}
	
	return s.rdb.SetEX(c, key, json, UserExpTime).Err()
}

func (s *UserStore) Delete(c context.Context, userID int64) { 
	key := fmt.Sprintf("user-%d", userID)

	s.rdb.Del(c, key)
}