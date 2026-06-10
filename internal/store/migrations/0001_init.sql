CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         CITEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    api_key_hash  TEXT NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_idx ON sessions (user_id);
CREATE INDEX sessions_expires_idx ON sessions (expires_at);

CREATE TABLE accounts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email         CITEXT NOT NULL,
    aas_token_enc BYTEA NOT NULL,
    visibility    TEXT NOT NULL DEFAULT 'private' CHECK (visibility IN ('public', 'private')),
    status        TEXT NOT NULL DEFAULT 'active'  CHECK (status IN ('active', 'flagged', 'disabled')),
    last_used_at  TIMESTAMPTZ,
    failure_count INT NOT NULL DEFAULT 0,
    mint_count    BIGINT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (owner_id, email)
);

-- Serves the hot rotation query: oldest active account in a pool.
CREATE INDEX accounts_rotation_idx
    ON accounts (visibility, status, last_used_at ASC NULLS FIRST);
CREATE INDEX accounts_owner_idx ON accounts (owner_id);

CREATE TABLE mint_events (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    account_id  UUID REFERENCES accounts(id) ON DELETE SET NULL,
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    anonymous   BOOLEAN NOT NULL DEFAULT TRUE,
    success     BOOLEAN NOT NULL,
    error       TEXT,
    duration_ms INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX mint_events_created_idx ON mint_events (created_at DESC);
CREATE INDEX mint_events_account_idx ON mint_events (account_id, created_at DESC);
