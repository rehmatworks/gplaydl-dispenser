-- Device users: the Android app enrols itself and syncs tokens without ever
-- creating an email/password identity, so both columns become optional.
-- Postgres allows multiple NULLs under a UNIQUE constraint, which keeps the
-- existing users_email_key working for web users.
ALTER TABLE users
    ALTER COLUMN email DROP NOT NULL,
    ALTER COLUMN password_hash DROP NOT NULL,
    ADD COLUMN kind TEXT NOT NULL DEFAULT 'web' CHECK (kind IN ('web', 'device')),
    ADD COLUMN device_id TEXT UNIQUE,
    ADD COLUMN label TEXT,
    ADD COLUMN consent_version TEXT,
    ADD COLUMN last_seen_at TIMESTAMPTZ;

-- Short-lived codes that let a phone-enrolled user open the web dashboard
-- without a password.
CREATE TABLE pairing_codes (
    code        TEXT PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX pairing_codes_user_idx ON pairing_codes (user_id);
CREATE INDEX pairing_codes_expires_idx ON pairing_codes (expires_at);

ALTER TABLE accounts
    ADD COLUMN source TEXT NOT NULL DEFAULT 'web' CHECK (source IN ('web', 'app')),
    ADD COLUMN consent_version TEXT,
    ADD COLUMN shared_at TIMESTAMPTZ,
    ADD COLUMN last_synced_at TIMESTAMPTZ;

-- Accounts that were already public predate consent tracking.
UPDATE accounts SET shared_at = created_at WHERE visibility = 'public';
