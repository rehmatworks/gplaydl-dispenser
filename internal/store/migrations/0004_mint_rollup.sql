-- One row per dispense grew without bound while only ever being read back as
-- three aggregate numbers and a 24-hour chart. Hourly buckets serve those
-- reads exactly and cap the table at 24 rows a day.
CREATE TABLE mint_stats_hourly (
    hour     TIMESTAMPTZ PRIMARY KEY,
    success  BIGINT NOT NULL DEFAULT 0,
    failures BIGINT NOT NULL DEFAULT 0
);

-- Lifetime totals, one row forever. Kept separate from the hourly buckets so
-- the public counter is a single-row read and survives bucket retention, and
-- so it cannot fall the way sum(accounts.mint_count) does when an account is
-- deleted.
CREATE TABLE mint_totals (
    id       BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
    success  BIGINT NOT NULL DEFAULT 0,
    failures BIGINT NOT NULL DEFAULT 0
);

-- Carry the existing history over so the chart is not blank after deploy.
INSERT INTO mint_stats_hourly (hour, success, failures)
SELECT date_trunc('hour', created_at),
       count(*) FILTER (WHERE success),
       count(*) FILTER (WHERE NOT success)
FROM mint_events
WHERE created_at > now() - interval '90 days'
GROUP BY 1;

-- Seed successes from the number the site already displays so the public
-- counter does not visibly jump on deploy.
INSERT INTO mint_totals (id, success, failures)
VALUES (TRUE,
        (SELECT coalesce(sum(mint_count), 0) FROM accounts),
        (SELECT count(*) FROM mint_events WHERE NOT success));

-- mint_events is intentionally left in place. Deploys run migrations on start
-- and roll the binary back on a failed health check, and the previous binary
-- still reads this table; dropping it here would turn a rollback into 500s on
-- the stats endpoints. It stops being written from this release onward and is
-- dropped in a later migration once this one has proven itself.
