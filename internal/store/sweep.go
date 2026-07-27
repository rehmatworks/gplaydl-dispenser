package store

import (
	"context"
	"log/slog"
	"time"
)

// mintStatsRetention keeps far more history than the 24-hour dashboard needs,
// because hourly buckets are cheap: a year is under 9k rows.
const mintStatsRetention = "90 days"

const sweepInterval = time.Hour

// Sweep deletes rows nothing can reach any more. Expired sessions and tokens
// are only removed on explicit logout or reuse otherwise, so without this they
// accumulate for the life of the deployment.
//
// Each statement stands alone; one failing should not hold back the others.
func (s *Store) Sweep(ctx context.Context, log *slog.Logger) {
	statements := []struct {
		what string
		sql  string
	}{
		{"mint_stats_hourly", `DELETE FROM mint_stats_hourly
			WHERE hour < now() - interval '` + mintStatsRetention + `'`},
		{"sessions", `DELETE FROM sessions WHERE expires_at < now()`},
		{"email_tokens", `DELETE FROM email_tokens WHERE expires_at < now()`},
		{"pairing_codes", `DELETE FROM pairing_codes
			WHERE expires_at < now() OR consumed_at IS NOT NULL`},
	}

	for _, st := range statements {
		tag, err := s.pool.Exec(ctx, st.sql)
		if err != nil {
			log.Error("sweep", "table", st.what, "err", err)
			continue
		}
		if n := tag.RowsAffected(); n > 0 {
			log.Info("swept expired rows", "table", st.what, "rows", n)
		}
	}
}

// StartSweeper runs Sweep once at startup and then hourly, until ctx is done.
func (s *Store) StartSweeper(ctx context.Context, log *slog.Logger) {
	go func() {
		ticker := time.NewTicker(sweepInterval)
		defer ticker.Stop()
		for {
			// Bound each pass so a stuck delete cannot wedge the loop.
			runCtx, cancel := context.WithTimeout(ctx, time.Minute)
			s.Sweep(runCtx, log)
			cancel()

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}
