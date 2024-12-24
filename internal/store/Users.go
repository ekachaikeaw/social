package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

type password struct {
	text *string
	hash []byte
}

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(c context.Context, u *User, tx *sql.Tx) error {
	query := `
		INSERT INTO users (username, email, password, role_id)	
		VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4))
		RETURNING id, created_at;
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	role := u.Role.Name
	if role == "" {
		role = "user"
	} 	

	err := tx.QueryRowContext(c, query, u.Username, u.Email, u.Password.hash, role).
		Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}
	return nil
}

func (s *UserStore) GetByID(c context.Context, userID int64) (*User, error) {
	query := `
		SELECT users.id, username, email, password, created_at, roles.* 
		FROM users
		JOIN roles ON users.role_id = roles.id
		WHERE users.id = $1 AND is_active = true;	
	`

	var user User

	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(c, query, userID).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password.hash,
			&user.CreatedAt,
			&user.Role.ID,
			&user.Role.Name,
			&user.Role.Level,
			&user.Role.Description,
		)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) GetByEmail(c context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, password, created_at
		FROM users
		WHERE email = $1 AND is_active = true;	
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	var user User
	err := s.db.QueryRowContext(c, query, email).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password.hash,
			&user.CreatedAt,
		)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) CreateAndInvite(c context.Context, u *User, token string, inviteExp time.Duration) error {
	return withTx(s.db, c, func(tx *sql.Tx) error {
		if err := s.Create(c, u, tx); err != nil {
			return err
		}

		if err := s.createUserInvitation(c, tx, token, inviteExp, u.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) Activate(c context.Context, token string) error {
	return withTx(s.db, c, func(tx *sql.Tx) error {
		// 1. find the user this token belong to
		user, err := s.getUserFromInvitation(c, tx, token)
		if err != nil {
			return err
		}

		// 2. update the user
		user.IsActive = true
		if err := s.update(c, tx, user); err != nil {
			return err
		}
		// 3. clean the invitations
		if err := s.deleteUserInvitation(c, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) Delete(c context.Context, userID int64) error {
	return withTx(s.db, c, func(tx *sql.Tx) error {
		if err := s.delete(c, tx, userID); err != nil {
			return err
		}

		if err := s.deleteUserInvitation(c, tx, userID); err != nil {
			return err
		}

		return nil
	})

}

func (s *UserStore) delete(c context.Context, tx *sql.Tx, userID int64) error {
	query := "DELETE FROM users WHERE id = $1;"

	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(c, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) createUserInvitation(c context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	query := `
		INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3);
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(c, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) getUserFromInvitation(c context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_active
		FROM users u
		JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2; 	
	`

	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	user := User{}
	err := tx.QueryRowContext(c, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) update(c context.Context, tx *sql.Tx, user *User) error {
	query := `
		UPDATE users SET username = $1, email = $2, is_active = $3
		WHERE users.id = $4;	
	`
	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(c, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) deleteUserInvitation(c context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1;`

	c, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(c, query, userID)
	if err != nil {
		return err
	}

	return nil
}
