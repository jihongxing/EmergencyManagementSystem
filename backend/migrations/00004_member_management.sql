-- +goose Up
CREATE TABLE ems.member_audit_events (
    id bigserial PRIMARY KEY,
    organization_id text NOT NULL REFERENCES ems.organizations(id),
    actor_member_id text NOT NULL REFERENCES ems.members(id),
    target_member_id text REFERENCES ems.members(id),
    action text NOT NULL CHECK (action IN ('member_created', 'member_roles_updated', 'member_status_updated', 'member_updated')),
    before_state jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(before_state) = 'object'),
    after_state jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(after_state) = 'object'),
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX member_audit_events_organization_time
    ON ems.member_audit_events (organization_id, occurred_at, id);

-- +goose Down
DROP TABLE ems.member_audit_events;
