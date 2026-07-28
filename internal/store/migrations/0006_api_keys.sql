-- The enrolment key on the users row belongs to the Android app itself.
-- Linking a gplaydl install mints an extra key here, one per machine, so a
-- laptop claiming a pairing code never locks the desktop out.
CREATE TABLE api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash     TEXT NOT NULL UNIQUE,
    label        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX api_keys_user_idx ON api_keys (user_id);
