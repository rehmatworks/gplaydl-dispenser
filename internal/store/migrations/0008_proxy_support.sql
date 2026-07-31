-- Administrative settings are opt-in. Operators promote trusted device users
-- manually in the database; enrolment never grants administrative access.
ALTER TABLE users
    ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE admin_settings (
    singleton          BOOLEAN PRIMARY KEY DEFAULT true CHECK (singleton),
    proxy_template_enc BYTEA,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO admin_settings (singleton) VALUES (true);

-- Concrete proxy assignments are encrypted because URLs commonly contain
-- usernames and passwords. Test state is safe metadata for dashboards.
ALTER TABLE accounts
    ADD COLUMN proxy_url_enc BYTEA,
    ADD COLUMN proxy_test_status TEXT
        CHECK (proxy_test_status IN ('passed', 'failed')),
    ADD COLUMN proxy_tested_at TIMESTAMPTZ;
