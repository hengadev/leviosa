-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pending_users (
    id TEXT PRIMARY KEY,
    email_hash TEXT NOT NULL,
    encrypted_email TEXT NOT NULL UNIQUE,
    password_hash TEXT,
	encrypted_birthdate  TEXT NOT NULL,
    encrypted_lastname TEXT NOT NULL,
    encrypted_firstname TEXT NOT NULL,
	encrypted_gender TEXT NOT NULL,
	encrypted_telephone TEXT NOT NULL UNIQUE,
    encrypted_created_at TEXT NOT NULL,
    encrypted_postal_code TEXT NOT NULL UNIQUE, 
    encrypted_city TEXT NOT NULL UNIQUE, 
    encrypted_address1 TEXT NOT NULL UNIQUE, 
    encrypted_address2 TEXT NOT NULL UNIQUE,
    encrypted_google_id TEXT,
    encrypted_apple_id TEXT 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE pending_users;
-- +goose StatementEnd
