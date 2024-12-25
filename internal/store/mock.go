package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct{}

func (s *MockUserStore) Create(context.Context, *User, *sql.Tx) error {
	return nil
}
func (s *MockUserStore) GetByID(context.Context, int64) (*User, error) {
	return &User{}, nil
}
func (s *MockUserStore) GetByEmail(context.Context, string) (*User, error) {
	return &User{}, nil
}
func (s *MockUserStore) CreateAndInvite(context.Context, *User, string, time.Duration) error {
	return nil
}
func (s *MockUserStore) Activate(context.Context, string) error {
	return nil
}
func (s *MockUserStore) Delete(context.Context, int64) error {
	return nil
}
