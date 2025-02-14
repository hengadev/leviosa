-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email_hash TEXT NOT NULL UNIQUE,
    encrypted_email TEXT NOT NULL UNIQUE,
    password_hash TEXT,
	encrypted_created_at  TEXT NOT NULL,
	encrypted_logged_in_at TEXT NOT NULL,
    role TEXT NOT NULL,
	encrypted_birthdate  TEXT NOT NULL,
    encrypted_lastname TEXT NOT NULL,
    encrypted_firstname TEXT NOT NULL,
	encrypted_gender TEXT NOT NULL,
	encrypted_telephone TEXT NOT NULL UNIQUE,
    encrypted_postal_code TEXT NOT NULL, 
    encrypted_city TEXT NOT NULL, 
    encrypted_address1 TEXT NOT NULL, 
    encrypted_address2 TEXT,
    encrypted_google_id TEXT,
    encrypted_apple_id TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
