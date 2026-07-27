package store

import (
	"context"
	"time"
)

// RecordMintOutcome books a dispense against the hourly bucket and the
// lifetime totals. Both are counter updates, so the tables stay a fixed size
// no matter how many mints run through them.
//
// Per-account bookkeeping (mint_count, failure_count, flagging) is separate
// and lives in RecordMintResult.
func (s *Store) RecordMintOutcome(ctx context.Context, success bool) error {
	ok, fail := 0, 1
	if success {
		ok, fail = 1, 0
	}
	// One round trip: this runs on the dispense path.
	_, err := s.pool.Exec(ctx, `
		WITH bump_totals AS (
			UPDATE mint_totals
			SET success = success + $1, failures = failures + $2
			WHERE id
		)
		INSERT INTO mint_stats_hourly AS h (hour, success, failures)
		VALUES (date_trunc('hour', now()), $1, $2)
		ON CONFLICT (hour) DO UPDATE
		SET success  = h.success + EXCLUDED.success,
		    failures = h.failures + EXCLUDED.failures`,
		ok, fail)
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
	// Contributors counts the distinct owners keeping the community pool alive.
	Contributors int64 `json:"contributors"`
	// SharedAccounts is the caller's own contribution to the public pool.
	SharedAccounts int64 `json:"sharedAccounts"`
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
			(SELECT coalesce(sum(success), 0) FROM mint_stats_hourly
				WHERE hour >= date_trunc('hour', now()) - interval '23 hours'),
			(SELECT coalesce(sum(failures), 0) FROM mint_stats_hourly
				WHERE hour >= date_trunc('hour', now()) - interval '23 hours'),
			(SELECT coalesce(max(success), 0) FROM mint_totals),
			(SELECT count(DISTINCT owner_id) FROM accounts WHERE visibility = 'public'),
			(SELECT count(*) FROM accounts WHERE visibility = 'public'
				AND ($1 = '' OR owner_id = $1::uuid))`,
		ownerID,
	).Scan(&st.PublicAccounts, &st.PrivateAccounts, &st.ActiveAccounts,
		&st.FlaggedAccounts, &st.Mints24h, &st.Failures24h, &st.TotalMints,
		&st.Contributors, &st.SharedAccounts)
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
	// generate_series keeps quiet hours in the result as explicit zeroes.
	rows, err := s.pool.Query(ctx, `
		SELECT
			h AS hour,
			coalesce(m.success, 0),
			coalesce(m.failures, 0)
		FROM generate_series(
			date_trunc('hour', now()) - interval '23 hours',
			date_trunc('hour', now()),
			interval '1 hour'
		) AS h
		LEFT JOIN mint_stats_hourly m ON m.hour = h
		ORDER BY hour`)
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
