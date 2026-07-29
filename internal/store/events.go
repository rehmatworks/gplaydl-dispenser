package store

import "context"

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
