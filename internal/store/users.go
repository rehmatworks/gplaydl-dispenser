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
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
}

const userCols = `id, email, email_verified_at IS NOT NULL, created_at`

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

func (s *Store) CreateUser(ctx context.Context, email, passwordHash, apiKeyHash string, verified bool) (*User, error) {
	u := &User{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, api_key_hash, email_verified_at)
		VALUES ($1, $2, $3, CASE WHEN $4 THEN now() END)
		RETURNING `+userCols,
		email, passwordHash, apiKeyHash, verified,
	).Scan(&u.ID, &u.Email, &u.EmailVerified, &u.CreatedAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	return u, nil
}

func (s *Store) UserByEmail(ctx context.Context, email string) (*User, string, error) {
	u := &User{}
	var passwordHash string
	err := s.pool.QueryRow(ctx, `
		SELECT `+userCols+`, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.EmailVerified, &u.CreatedAt, &passwordHash)
	if err != nil {
		return nil, "", wrapErr(err)
	}
	return u, passwordHash, nil
}

func (s *Store) UserByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := s.pool.QueryRow(ctx, `
		SELECT `+userCols+` FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.EmailVerified, &u.CreatedAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	return u, nil
}

func (s *Store) UserByAPIKeyHash(ctx context.Context, keyHash string) (*User, error) {
	u := &User{}
	err := s.pool.QueryRow(ctx, `
		SELECT `+userCols+` FROM users WHERE api_key_hash = $1`, keyHash,
	).Scan(&u.ID, &u.Email, &u.EmailVerified, &u.CreatedAt)
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

func (s *Store) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET email_verified_at = COALESCE(email_verified_at, now()) WHERE id = $1`,
		userID)
	return err
}

func (s *Store) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2 WHERE id = $1`, userID, passwordHash)
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
		SELECT u.id, u.email, u.email_verified_at IS NOT NULL, u.created_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()`,
		tokenHash,
	).Scan(&u.ID, &u.Email, &u.EmailVerified, &u.CreatedAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	return u, nil
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

// DeleteUserSessions revokes every session, e.g. after a password reset.
func (s *Store) DeleteUserSessions(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

// --- Email tokens (verification + password reset) ---

func (s *Store) CreateEmailToken(ctx context.Context, tokenHash, userID, purpose string, ttl time.Duration) error {
	// One outstanding token per user+purpose; a new request replaces the old.
	_, err := s.pool.Exec(ctx, `
		DELETE FROM email_tokens WHERE user_id = $1 AND purpose = $2`,
		userID, purpose)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO email_tokens (token_hash, user_id, purpose, expires_at)
		VALUES ($1, $2, $3, now() + $4)`,
		tokenHash, userID, purpose, ttl)
	return err
}

// ConsumeEmailToken validates and deletes a token in one step (single use).
func (s *Store) ConsumeEmailToken(ctx context.Context, tokenHash, purpose string) (string, error) {
	var userID string
	err := s.pool.QueryRow(ctx, `
		DELETE FROM email_tokens
		WHERE token_hash = $1 AND purpose = $2 AND expires_at > now()
		RETURNING user_id`,
		tokenHash, purpose,
	).Scan(&userID)
	if err != nil {
		return "", wrapErr(err)
	}
	return userID, nil
}
