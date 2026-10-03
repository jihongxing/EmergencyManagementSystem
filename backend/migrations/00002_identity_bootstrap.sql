-- +goose Up
CREATE TABLE ems.organizations (
    id text PRIMARY KEY CHECK (id ~ '^org_[A-Za-z0-9_-]+$'),
    kind text NOT NULL CHECK (kind IN ('enterprise', 'department')),
    name text NOT NULL CHECK (length(btrim(name)) BETWEEN 1 AND 200),
    external_key text NOT NULL CHECK (length(btrim(external_key)) BETWEEN 1 AND 200),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'closed')),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (kind, external_key)
);

CREATE TABLE ems.members (
    id text PRIMARY KEY CHECK (id ~ '^mem_[A-Za-z0-9_-]+$'),
    user_id text NOT NULL CHECK (user_id ~ '^usr_[A-Za-z0-9_-]+$'),
    organization_id text NOT NULL REFERENCES ems.organizations(id),
    login_id text NOT NULL CHECK (length(btrim(login_id)) BETWEEN 1 AND 200),
    status text NOT NULL CHECK (status IN ('pending', 'active', 'suspended', 'revoked')),
    roles jsonb NOT NULL CHECK (jsonb_typeof(roles) = 'array' AND jsonb_array_length(roles) > 0),
    is_first_admin boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, login_id)
);

CREATE UNIQUE INDEX members_one_current_first_admin
    ON ems.members (organization_id)
    WHERE is_first_admin AND status <> 'revoked';

CREATE TABLE ems.organization_materials (
    id bigserial PRIMARY KEY,
    organization_id text NOT NULL REFERENCES ems.organizations(id),
    source_id text NOT NULL CHECK (length(btrim(source_id)) > 0),
    verified boolean NOT NULL,
    verified_at timestamptz,
    verified_by text NOT NULL CHECK (length(btrim(verified_by)) > 0),
    recorded_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ems.identity_audit_events (
    id bigserial PRIMARY KEY,
    organization_id text NOT NULL REFERENCES ems.organizations(id),
    actor_id text NOT NULL CHECK (length(btrim(actor_id)) > 0),
    action text NOT NULL CHECK (action IN ('organization_bootstrap', 'first_admin_replaced')),
    member_id text,
    details jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(details) = 'object'),
    occurred_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE ems.identity_audit_events;
DROP TABLE ems.organization_materials;
DROP TABLE ems.members;
DROP TABLE ems.organizations;
