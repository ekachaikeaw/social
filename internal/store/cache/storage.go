package cache

import (
	"context"

	"github.com/ekachaikeaw/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
		Delete(context.Context, int64)
	}
}

func NewCacheStorage(rdb *redis.Client) Storage {
	return Storage{
		&UserStore{rdb: rdb,},
	}
}