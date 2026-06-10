package store

import (
	"context"
	"time"
)

type MintEvent struct {
	AccountID  string
	UserID     string
	Anonymous  bool
	Success    bool
	Error      string
	DurationMS int
}

func (s *Store) RecordMintEvent(ctx context.Context, e MintEvent) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO mint_events (account_id, user_id, anonymous, success, error, duration_ms)
		VALUES (NULLIF($1, '')::uuid, NULLIF($2, '')::uuid, $3, $4, NULLIF($5, ''), $6)`,
		e.AccountID, e.UserID, e.Anonymous, e.Success, e.Error, e.DurationMS)
	return err
}

type PoolStats struct {
	PublicAccounts  int64 `json:"publicAccounts"`
	PrivateAccounts int64 `json:"privateAccounts"`
	ActiveAccounts  int64 `json:"activeAccounts"`
	FlaggedAccounts int64 `json:"flaggedAccounts"`
	Mints24h        int64 `json:"mints24h"`
	Failures24h     int64 `json:"failures24h"`
	TotalMints      int64 `json:"totalMints"`
}

// Stats returns pool-wide numbers; when ownerID is non-empty the account
// counters are scoped to that user while mint numbers stay pool-wide.
func (s *Store) Stats(ctx context.Context, ownerID string) (*PoolStats, error) {
	st := &PoolStats{}
	err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM accounts WHERE visibility = 'public' AND status = 'active'),
			(SELECT count(*) FROM accounts WHERE visibility = 'private'
				AND ($1 = '' OR owner_id = $1::uuid)),
			(SELECT count(*) FROM accounts WHERE status = 'active'
				AND ($1 = '' OR owner_id = $1::uuid)),
			(SELECT count(*) FROM accounts WHERE status = 'flagged'
				AND ($1 = '' OR owner_id = $1::uuid)),
			(SELECT count(*) FROM mint_events WHERE success AND created_at > now() - interval '24 hours'),
			(SELECT count(*) FROM mint_events WHERE NOT success AND created_at > now() - interval '24 hours'),
			(SELECT coalesce(sum(mint_count), 0) FROM accounts)`,
		ownerID,
	).Scan(&st.PublicAccounts, &st.PrivateAccounts, &st.ActiveAccounts,
		&st.FlaggedAccounts, &st.Mints24h, &st.Failures24h, &st.TotalMints)
	if err != nil {
		return nil, err
	}
	return st, nil
}

type MintBucket struct {
	Hour     time.Time `json:"hour"`
	Success  int64     `json:"success"`
	Failures int64     `json:"failures"`
}

// MintTimeline returns hourly mint counts for the last 24 hours.
func (s *Store) MintTimeline(ctx context.Context) ([]MintBucket, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			date_trunc('hour', h) AS hour,
			coalesce(count(e.id) FILTER (WHERE e.success), 0),
			coalesce(count(e.id) FILTER (WHERE NOT e.success), 0)
		FROM generate_series(
			date_trunc('hour', now()) - interval '23 hours',
			date_trunc('hour', now()),
			interval '1 hour'
		) AS h
		LEFT JOIN mint_events e
			ON date_trunc('hour', e.created_at) = h
		GROUP BY hour ORDER BY hour`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buckets := []MintBucket{}
	for rows.Next() {
		var b MintBucket
		if err := rows.Scan(&b.Hour, &b.Success, &b.Failures); err != nil {
			return nil, err
		}
		buckets = append(buckets, b)
	}
	return buckets, rows.Err()
}
