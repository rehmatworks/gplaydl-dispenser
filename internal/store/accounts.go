package store

import (
	"context"
	"time"
)

type Account struct {
	ID           string     `json:"id"`
	OwnerID      string     `json:"ownerId"`
	Email        string     `json:"email"`
	AASTokenEnc  []byte     `json:"-"`
	Visibility   string     `json:"visibility"`
	Status       string     `json:"status"`
	LastUsedAt   *time.Time `json:"lastUsedAt"`
	FailureCount int        `json:"failureCount"`
	MintCount    int64      `json:"mintCount"`
	CreatedAt    time.Time  `json:"createdAt"`
}

const accountCols = `id, owner_id, email, aas_token_enc, visibility, status,
	last_used_at, failure_count, mint_count, created_at`

func scanAccount(row interface{ Scan(...any) error }) (*Account, error) {
	a := &Account{}
	err := row.Scan(&a.ID, &a.OwnerID, &a.Email, &a.AASTokenEnc, &a.Visibility,
		&a.Status, &a.LastUsedAt, &a.FailureCount, &a.MintCount, &a.CreatedAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	return a, nil
}

func (s *Store) CreateAccount(ctx context.Context, ownerID, email string, aasTokenEnc []byte, visibility string) (*Account, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO accounts (owner_id, email, aas_token_enc, visibility)
		VALUES ($1, $2, $3, $4)
		RETURNING `+accountCols,
		ownerID, email, aasTokenEnc, visibility)
	return scanAccount(row)
}

func (s *Store) AccountsByOwner(ctx context.Context, ownerID string) ([]*Account, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+accountCols+` FROM accounts
		WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (s *Store) AccountByID(ctx context.Context, id, ownerID string) (*Account, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+accountCols+` FROM accounts WHERE id = $1 AND owner_id = $2`,
		id, ownerID)
	return scanAccount(row)
}

func (s *Store) UpdateAccount(ctx context.Context, id, ownerID string, visibility, status *string) (*Account, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE accounts SET
			visibility = COALESCE($3, visibility),
			status     = COALESCE($4, status),
			-- manual re-enable gives the account a clean slate
			failure_count = CASE WHEN $4 = 'active' THEN 0 ELSE failure_count END,
			updated_at = now()
		WHERE id = $1 AND owner_id = $2
		RETURNING `+accountCols,
		id, ownerID, visibility, status)
	return scanAccount(row)
}

func (s *Store) DeleteAccount(ctx context.Context, id, ownerID string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM accounts WHERE id = $1 AND owner_id = $2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// NextAccount atomically claims the least-recently-used active account.
// FOR UPDATE SKIP LOCKED makes concurrent dispenses pick distinct accounts
// without blocking each other, and the rotation survives restarts.
//
// pool semantics:
//   - ownerID == ""  → public pool only (anonymous dispense)
//   - ownerID != ""  → that user's own accounts, plus the public pool when
//     includePublic is true
func (s *Store) NextAccount(ctx context.Context, ownerID string, includePublic bool) (*Account, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE accounts SET last_used_at = now(), updated_at = now()
		WHERE id = (
			SELECT id FROM accounts
			WHERE status = 'active'
			  AND (
			        ($1 = '' AND visibility = 'public')
			     OR ($1 <> '' AND owner_id = $1::uuid)
			     OR ($1 <> '' AND $2 AND visibility = 'public')
			  )
			ORDER BY last_used_at ASC NULLS FIRST
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING `+accountCols,
		ownerID, includePublic)
	return scanAccount(row)
}

const flagThreshold = 5

// RecordMintResult updates rotation counters and auto-flags an account after
// too many consecutive failures so dead credentials drop out of the pool.
func (s *Store) RecordMintResult(ctx context.Context, accountID string, success bool) error {
	if success {
		_, err := s.pool.Exec(ctx, `
			UPDATE accounts SET
				failure_count = 0,
				mint_count = mint_count + 1,
				status = CASE WHEN status = 'flagged' THEN 'active' ELSE status END,
				updated_at = now()
			WHERE id = $1`, accountID)
		return err
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE accounts SET
			failure_count = failure_count + 1,
			status = CASE
				WHEN failure_count + 1 >= $2 AND status = 'active' THEN 'flagged'
				ELSE status
			END,
			updated_at = now()
		WHERE id = $1`, accountID, flagThreshold)
	return err
}
