package cache

import (
	"context"

	"github.com/ekachaikeaw/social/internal/store"
)

type MockUserStore struct{}

func NewMockStore() Storage {
	return Storage{
		&MockUserStore{},
	}
}

func (s *MockUserStore) Get(context.Context, int64) (*store.User, error) {
	return &store.User{}, nil
}

func (s *MockUserStore) Set(context.Context, *store.User) error {
	return nil
}
func (s *MockUserStore) Delete(context.Context, int64) {
	
}
