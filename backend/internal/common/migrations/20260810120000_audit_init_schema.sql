-- +goose Up
-- +goose StatementBegin

-- Append-only audit log for sensitive admin/staff actions (role changes,
-- booking refunds/admin overrides, settings field updates). No update or
-- delete path is exposed by the Go port (internal/common/audit) — this
-- table is intentionally insert-only.
CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS audit.entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID NOT NULL,
    actor_role TEXT NOT NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_entries_target ON audit.entries(target_type, target_id);
CREATE INDEX idx_audit_entries_actor ON audit.entries(actor_id);
CREATE INDEX idx_audit_entries_created_at ON audit.entries(created_at);

COMMENT ON TABLE audit.entries IS 'Append-only audit trail for sensitive actions (role changes, refunds, admin overrides, settings changes). No update/delete path by design.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS audit.entries;
DROP SCHEMA IF EXISTS audit;

-- +goose StatementEnd
