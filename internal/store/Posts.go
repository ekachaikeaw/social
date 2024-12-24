package store

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	Version   int       `json:"version"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentCount int64 `json:"comment_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) GetUserFeed(c context.Context, followerID int64, fq PaginatedQuery) ([]PostWithMetadata, error) {
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
			count(c.id) AS comment_count, u.username
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON u.id = p.user_id
		JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
		WHERE 
			f.user_id = $1 AND 
			(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
			(p.tags @> $5 OR $5 = '{}') 
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + fq.Sort + ` 
		LIMIT $2 OFFSET $3;	
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	log.Println(pq.Array(fq.Tags))
	rows, err := s.db.QueryContext(c, query, followerID, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var feed []PostWithMetadata
	for rows.Next() {
		var p PostWithMetadata
		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.CommentCount,
			&p.User.Username,
		)
		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, nil
}

func (s *PostStore) Create(c context.Context, p *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at;
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		c,
		query,
		p.Content,
		p.Title,
		p.UserID,
		pq.Array(p.Tags),
	).Scan(
		&p.ID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetByID(c context.Context, postID int64) (*Post, error) {
	query := `
		SELECT id, content, title, tags, version, user_id, created_at, updated_at FROM posts WHERE id=$1;
	`

	var p Post
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(c, query, postID).
		Scan(
			&p.ID,
			&p.Content,
			&p.Title,
			pq.Array(&p.Tags),
			&p.Version,
			&p.UserID,
			&p.CreatedAt,
			&p.UpdatedAt,
		)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &p, nil
}

func (s *PostStore) Delete(c context.Context, id int64) error {
	query := `
		DELETE FROM posts 
		WHERE id=$1;
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(c, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostStore) Update(c context.Context, p *Post) error {
	query := `
		UPDATE posts
		SET title=$1, content=$2, version = version + 1
		WHERE id=$3 AND version=$4
		RETURNING version;	
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(c, query, p.Title, p.Content, p.ID, p.Version).Scan(&p.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}
