-- +goose Up
ALTER TABLE ems.members ADD COLUMN password_hash text;

CREATE TABLE ems.sessions (
    id text PRIMARY KEY CHECK (id ~ '^ses_[A-Za-z0-9_-]+$'),
    member_id text NOT NULL REFERENCES ems.members(id),
    client text NOT NULL CHECK (client IN ('web', 'flutter_android', 'flutter_ios')),
    access_token_hash text NOT NULL UNIQUE,
    refresh_token_hash text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    access_expires_at timestamptz NOT NULL,
    refresh_expires_at timestamptz NOT NULL,
    revoked_at timestamptz
);

CREATE TABLE ems.password_resets (
    id text PRIMARY KEY CHECK (id ~ '^rst_[A-Za-z0-9_-]+$'),
    member_id text NOT NULL REFERENCES ems.members(id),
    token_hash text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    revoked_at timestamptz
);

-- +goose Down
DROP TABLE ems.password_resets;
DROP TABLE ems.sessions;
ALTER TABLE ems.members DROP COLUMN password_hash;
