-- +goose Up
-- +goose StatementBegin

-- The pepper canary (issue 015; package internal/common/encryption/canary):
-- one row holding a fixed public constant encrypted through encx (its DEK
-- wrapped by the KMS KEK) and the constant's argon2id hash, which mixes in the
-- pepper. serve inserts it at startup when it is absent and never updates it;
-- the verify mode decrypts it and checks the hash to prove the database is
-- readable with the configured KMS key and pepper. Never delete or rewrite it.
CREATE SCHEMA IF NOT EXISTS encryption;

CREATE TABLE IF NOT EXISTS encryption.canary (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    value_encrypted BYTEA NOT NULL,
    value_hash_secure TEXT NOT NULL,
    dek_encrypted BYTEA NOT NULL,
    key_version INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE encryption.canary IS 'Pepper canary: a fixed constant encrypted with the KMS KEK and hashed with the pepper. Written once by serve, checked by the verify mode. Never delete or rewrite.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS encryption.canary;
DROP SCHEMA IF EXISTS encryption;

-- +goose StatementEnd
