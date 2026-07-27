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
	CreatedAt time.Time `json:"createdAt"`
	// Kind is "web" for email/password users and "device" for app enrolments.
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

const userCols = `id, created_at, kind, coalesce(label, '')`

func scanUser(row interface{ Scan(...any) error }, extra ...any) (*User, error) {
	u := &User{}
	dest := append([]any{&u.ID, &u.CreatedAt, &u.Kind, &u.Label}, extra...)
	if err := row.Scan(dest...); err != nil {
		return nil, wrapErr(err)
	}
	return u, nil
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

// CreateDeviceUser registers an Android install as its own identity. There is
// no email to verify, so the row is created pre-verified and can add accounts
// immediately.
func (s *Store) CreateDeviceUser(ctx context.Context, deviceID, label, apiKeyHash, consentVersion string) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (kind, device_id, label, api_key_hash, consent_version,
		                   email_verified_at, last_seen_at)
		VALUES ('device', $1, $2, $3, $4, now(), now())
		RETURNING `+userCols,
		deviceID, label, apiKeyHash, consentVersion)
	return scanUser(row)
}

// ReissueDeviceKey rotates the API key of an existing enrolment. Re-installing
// the app with the same device id recovers the identity instead of orphaning
// every account it had shared.
func (s *Store) ReissueDeviceKey(ctx context.Context, deviceID, label, apiKeyHash, consentVersion string) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE users SET
			api_key_hash = $3,
			label = COALESCE(NULLIF($2, ''), label),
			consent_version = COALESCE(NULLIF($4, ''), consent_version),
			last_seen_at = now()
		WHERE device_id = $1 AND kind = 'device'
		RETURNING `+userCols,
		deviceID, label, apiKeyHash, consentVersion)
	return scanUser(row)
}

func (s *Store) UserByID(ctx context.Context, id string) (*User, error) {
	return scanUser(s.pool.QueryRow(ctx, `SELECT `+userCols+` FROM users WHERE id = $1`, id))
}

func (s *Store) UserByAPIKeyHash(ctx context.Context, keyHash string) (*User, error) {
	return scanUser(s.pool.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE api_key_hash = $1`, keyHash))
}

// TouchUser records app activity so the dashboard can show when a device last
// checked in. Best-effort: callers ignore the error.
func (s *Store) TouchUser(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET last_seen_at = now() WHERE id = $1`, userID)
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
	return scanUser(s.pool.QueryRow(ctx, `
		SELECT u.id, u.created_at, u.kind, coalesce(u.label, '')
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()`,
		tokenHash))
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

// --- Pairing codes (app -> dashboard hand-off) ---

func (s *Store) CreatePairingCode(ctx context.Context, code, userID string, ttl time.Duration) error {
	// One outstanding code per device; asking again replaces the old one.
	if _, err := s.pool.Exec(ctx,
		`DELETE FROM pairing_codes WHERE user_id = $1`, userID); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO pairing_codes (code, user_id, expires_at)
		VALUES ($1, $2, now() + $3)`,
		code, userID, ttl)
	return err
}

// ConsumePairingCode redeems a code exactly once and returns its owner.
func (s *Store) ConsumePairingCode(ctx context.Context, code string) (string, error) {
	var userID string
	err := s.pool.QueryRow(ctx, `
		UPDATE pairing_codes SET consumed_at = now()
		WHERE code = $1 AND consumed_at IS NULL AND expires_at > now()
		RETURNING user_id`,
		code,
	).Scan(&userID)
	if err != nil {
		return "", wrapErr(err)
	}
	return userID, nil
}
