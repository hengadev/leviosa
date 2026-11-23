-- +goose Up
-- +goose StatementBegin

-- Web Push subscriptions (issue 044, ADR-0020): the devices on which a user
-- (the practitioner) receives practitioner notifications. Several per user.
-- The endpoint is a bearer capability, so it is stored encrypted with its
-- keys (encx, ADR-0015); endpoint_hash (encx basic hash) finds a device again
-- on unsubscribe and keeps one row per device. No free-text column
-- (ADR-0019).
--
-- Compatible with the previous release: a new schema and table it neither
-- reads nor writes. Deleting a user takes their subscriptions with them.
CREATE SCHEMA IF NOT EXISTS notification;

CREATE TABLE notification.push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    endpoint_hash VARCHAR(64) NOT NULL,
    endpoint_encrypted BYTEA NOT NULL,
    p256dh_encrypted BYTEA NOT NULL,
    auth_encrypted BYTEA NOT NULL,
    user_agent TEXT NOT NULL DEFAULT '',
    dek_encrypted BYTEA NOT NULL,
    key_version INTEGER NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT push_subscriptions_endpoint_hash_key UNIQUE (endpoint_hash)
);

CREATE INDEX idx_push_subscriptions_user_id ON notification.push_subscriptions (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS notification.push_subscriptions;
DROP SCHEMA IF EXISTS notification;

-- +goose StatementEnd
