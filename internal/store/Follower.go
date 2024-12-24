package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Follower struct {
	FollowerID int64 `json:"follower_id"`
	UserID     int64 `json:"user_id"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(c context.Context, followerID, userID int64) error {
	query := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2);
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(c, query, userID, followerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		} 
	}

	return nil
}

func (s *FollowerStore) Unfollow(c context.Context, followerID, userID int64) error {
	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2;	
	`

	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(c, query, userID, followerID)
	if err != nil {
		return err
	}

	return nil
}