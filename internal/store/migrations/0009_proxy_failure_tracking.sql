-- Keep runtime proxy health separate from Google credential failures so a
-- broken route cannot incorrectly flag an otherwise valid account.
ALTER TABLE accounts
    ADD COLUMN proxy_failure_count INT NOT NULL DEFAULT 0,
    ADD COLUMN last_proxy_failure_at TIMESTAMPTZ;
