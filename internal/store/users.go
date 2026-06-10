package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("already exists")
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func wrapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicate
	}
	return err
}

func (s *Store) CreateUser(ctx context.Context, email, passwordHash, apiKeyHash string) (*User, error) {
	u := &User{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, api_key_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, created_at`,
		email, passwordHash, apiKeyHash,
	).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	return u, nil
}

func (s *Store) UserByEmail(ctx context.Context, email string) (*User, string, error) {
	u := &User{}
	var passwordHash string
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, created_at, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.CreatedAt, &passwordHash)
	if err != nil {
		return nil, "", wrapErr(err)
	}
	return u, passwordHash, nil
}

func (s *Store) UserByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	return u, nil
}

func (s *Store) UserByAPIKeyHash(ctx context.Context, keyHash string) (*User, error) {
	u := &User{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, created_at FROM users WHERE api_key_hash = $1`, keyHash,
	).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	return u, nil
}

func (s *Store) RotateAPIKey(ctx context.Context, userID, newKeyHash string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET api_key_hash = $2 WHERE id = $1`, userID, newKeyHash)
	return err
}

// --- Sessions ---

func (s *Store) CreateSession(ctx context.Context, tokenHash, userID string, ttl time.Duration) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2, now() + $3)`,
		tokenHash, userID, ttl)
	return err
}

func (s *Store) UserBySession(ctx context.Context, tokenHash string) (*User, error) {
	u := &User{}
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.created_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()`,
		tokenHash,
	).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	return u, nil
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}
