package store

import (
	"context"
	"time"
)

type Account struct {
	ID                 string     `json:"id"`
	OwnerID            string     `json:"ownerId"`
	Email              string     `json:"email"`
	AASTokenEnc        []byte     `json:"-"`
	Visibility         string     `json:"visibility"`
	Status             string     `json:"status"`
	LastUsedAt         *time.Time `json:"lastUsedAt"`
	FailureCount       int        `json:"failureCount"`
	MintCount          int64      `json:"mintCount"`
	CreatedAt          time.Time  `json:"createdAt"`
	Source             string     `json:"source"`
	SharedAt           *time.Time `json:"sharedAt"`
	LastSyncedAt       *time.Time `json:"lastSyncedAt"`
	ProxyURLEnc        []byte     `json:"-"`
	ProxyConfigured    bool       `json:"proxyConfigured"`
	ProxyTestStatus    string     `json:"proxyTestStatus,omitempty"`
	ProxyTestedAt      *time.Time `json:"proxyTestedAt,omitempty"`
	ProxyFailureCount  int        `json:"proxyFailureCount"`
	LastProxyFailureAt *time.Time `json:"lastProxyFailureAt,omitempty"`
}

const accountCols = `id, owner_id, email, aas_token_enc, visibility, status,
	last_used_at, failure_count, mint_count, created_at, source, shared_at, last_synced_at,
	proxy_url_enc, coalesce(proxy_test_status, ''), proxy_tested_at,
	proxy_failure_count, last_proxy_failure_at`

func scanAccount(row interface{ Scan(...any) error }) (*Account, error) {
	a := &Account{}
	err := row.Scan(&a.ID, &a.OwnerID, &a.Email, &a.AASTokenEnc, &a.Visibility,
		&a.Status, &a.LastUsedAt, &a.FailureCount, &a.MintCount, &a.CreatedAt,
		&a.Source, &a.SharedAt, &a.LastSyncedAt, &a.ProxyURLEnc,
		&a.ProxyTestStatus, &a.ProxyTestedAt, &a.ProxyFailureCount,
		&a.LastProxyFailureAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	a.ProxyConfigured = len(a.ProxyURLEnc) > 0
	return a, nil
}

// UpsertAccount stores a Google account, replacing the token if the address is
// already registered. The app re-syncs whenever it re-mints, so an insert-only
// path would reject every refresh.
//
// A Google address exists once across the whole dispenser. Completing sign-in
// proves control of it, so a sync from a different device takes the account
// over rather than adding a second copy: one Google account would otherwise
// appear as several, and every previous owner would keep a live token for an
// address they may no longer control.
//
// Every account is private to its owner. The visibility column stays 'private'
// for compatibility with the still-present schema.
func (s *Store) UpsertAccount(ctx context.Context, ownerID, email string, aasTokenEnc []byte, source string) (*Account, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO accounts (owner_id, email, aas_token_enc, visibility, source,
		                      last_synced_at)
		VALUES ($1, $2, $3, 'private', $4, now())
		ON CONFLICT (email) DO UPDATE SET
			owner_id        = EXCLUDED.owner_id,
			aas_token_enc   = EXCLUDED.aas_token_enc,
			source          = EXCLUDED.source,
			-- a fresh token deserves a clean slate in the rotation
			status          = 'active',
			failure_count   = 0,
			last_synced_at  = now(),
			updated_at      = now()
		RETURNING `+accountCols,
		ownerID, email, aasTokenEnc, source)
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

// SetAccountProxyIfMissing installs the first generated assignment and leaves
// an existing assignment untouched when concurrent syncs race.
func (s *Store) SetAccountProxyIfMissing(ctx context.Context, accountID string, proxyURLEnc []byte, testStatus string, testedAt time.Time) (bool, error) {
	failureCount := 0
	var lastFailureAt *time.Time
	if testStatus == "failed" {
		failureCount = 1
		lastFailureAt = &testedAt
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE accounts SET
			proxy_url_enc = $2,
			proxy_test_status = $3,
			proxy_tested_at = $4,
			proxy_failure_count = $5,
			last_proxy_failure_at = $6,
			updated_at = now()
		WHERE id = $1 AND proxy_url_enc IS NULL`,
		accountID, proxyURLEnc, testStatus, testedAt, failureCount, lastFailureAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// AccountsNeedingProxyBackfill returns accounts without an assignment or
// whose current assignment has recorded a failed connection. Healthy proxies
// are intentionally left stable.
func (s *Store) AccountsNeedingProxyBackfill(ctx context.Context) ([]*Account, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+accountCols+` FROM accounts
		WHERE proxy_url_enc IS NULL
		   OR proxy_test_status = 'failed'
		   OR proxy_failure_count > 0
		ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

// ReplaceAccountProxy installs a freshly expanded assignment during an
// administrator-initiated backfill.
func (s *Store) ReplaceAccountProxy(ctx context.Context, accountID string, proxyURLEnc []byte, testStatus string, testedAt time.Time) error {
	failureCount := 0
	var lastFailureAt *time.Time
	if testStatus == "failed" {
		failureCount = 1
		lastFailureAt = &testedAt
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE accounts SET
			proxy_url_enc = $2,
			proxy_test_status = $3,
			proxy_tested_at = $4,
			proxy_failure_count = $5,
			last_proxy_failure_at = $6,
			updated_at = now()
		WHERE id = $1`,
		accountID, proxyURLEnc, testStatus, testedAt, failureCount, lastFailureAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// RecordProxyResult tracks consecutive connection failures independently from
// credential failures. A successful proxied connection immediately restores
// the account's proxy health and resets the counter.
func (s *Store) RecordProxyResult(ctx context.Context, accountID string, success bool) (int, error) {
	var count int
	if success {
		err := s.pool.QueryRow(ctx, `
			UPDATE accounts SET
				proxy_failure_count = 0,
				proxy_test_status = 'passed',
				updated_at = now()
			WHERE id = $1
			RETURNING proxy_failure_count`, accountID).Scan(&count)
		return count, wrapErr(err)
	}
	err := s.pool.QueryRow(ctx, `
		UPDATE accounts SET
			proxy_failure_count = proxy_failure_count + 1,
			proxy_test_status = 'failed',
			last_proxy_failure_at = now(),
			updated_at = now()
		WHERE id = $1
		RETURNING proxy_failure_count`, accountID).Scan(&count)
	return count, wrapErr(err)
}

// NextAccount atomically claims the caller's least-recently-used active
// account. FOR UPDATE SKIP LOCKED makes concurrent dispenses pick distinct
// accounts without blocking each other, and the rotation survives restarts.
//
// Every account is private to its owner, so a dispense only ever draws from
// the accounts that owner added.
func (s *Store) NextAccount(ctx context.Context, ownerID string) (*Account, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE accounts SET last_used_at = now(), updated_at = now()
		WHERE id = (
			SELECT id FROM accounts
			WHERE status = 'active' AND owner_id = $1::uuid
			ORDER BY last_used_at ASC NULLS FIRST
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING `+accountCols,
		ownerID)
	return scanAccount(row)
}

// NextAccountForEmail claims one specific account of an owner, identified by
// its Google address. This is how someone with several accounts downloads as a
// particular one: `--email me@gmail.com` pins the rotation to that account.
func (s *Store) NextAccountForEmail(ctx context.Context, ownerID, email string) (*Account, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE accounts SET last_used_at = now(), updated_at = now()
		WHERE id = (
			SELECT id FROM accounts
			WHERE owner_id = $1::uuid AND email = $2 AND status <> 'disabled'
			LIMIT 1
			FOR UPDATE
		)
		RETURNING `+accountCols,
		ownerID, email)
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
