package store

import (
	"context"
	"database/sql"

)

type Comment struct {
	ID         int64  `json:"id"`
	PostID     int64  `json:"post_id"`
	UserID     int64  `json:"user_id"`
	Content    string `json:"content"`
	User       User   `json:"user"`
	CreatedAt string `json:"created_at"`
}

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) GetByPostID(c context.Context, id int64) ([]Comment, error) {
	query := `
        SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.username, users.id FROM comments c 
        JOIN users ON users.id = c.user_id 
        WHERE c.post_id = $1
        ORDER BY c.created_at DESC; 
    `

	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(c, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		c.User = User{}	
		err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.CreatedAt, &c.User.Username, &c.User.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (s *CommentStore) Create(c context.Context, comment *Comment) error{
	query := `
		INSERT INTO comments (user_id, post_id, content)
		VALUES ($1, $2, $3) 
		RETURNING id, created_at;	
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(c, query, comment.UserID, comment.PostID, comment.Content).
		Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		return err
	}
	
	return nil
}