-- Uniqueness was per owner, so the same Google address could be registered by
-- several people at once. The rotation then treated one Google account as
-- several, overstating how much cover the pool really had, and every past owner
-- kept a usable token for it. Signing in proves control of the address, so the
-- address belongs to whoever did that most recently, and only to them.

-- Rank duplicates by when each row was last authenticated. Rows created
-- through the old web flow never synced, so fall back to their creation time.
CREATE TEMP TABLE account_dupes ON COMMIT DROP AS
SELECT id,
       email,
       mint_count,
       row_number() OVER (
           PARTITION BY email
           ORDER BY coalesce(last_synced_at, created_at) DESC, created_at DESC, id
       ) AS seq
FROM accounts;

-- The duplicates were all the same Google account, so their dispense counts
-- belong to the row that survives rather than disappearing from pool totals.
UPDATE accounts SET mint_count = accounts.mint_count + folded.extra
FROM (
    SELECT winner.id, sum(stale.mint_count) AS extra
    FROM account_dupes winner
    JOIN account_dupes stale ON stale.email = winner.email AND stale.seq > 1
    WHERE winner.seq = 1
    GROUP BY winner.id
) AS folded
WHERE accounts.id = folded.id;

DELETE FROM accounts
WHERE id IN (SELECT id FROM account_dupes WHERE seq > 1);

ALTER TABLE accounts ADD CONSTRAINT accounts_email_key UNIQUE (email);

-- accounts_owner_id_email_key is now implied by the constraint above, but it is
-- left in place deliberately. A failed health check rolls the binary back while
-- this migration stays applied, and the previous binary's upsert names
-- (owner_id, email) as its conflict target; dropping it would turn a rollback
-- into a 500 on every account sync. It also still serves the per-owner lookup
-- behind ?email=. A later migration can drop it once this one has settled.
