-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS unverified_users (
    email_hash TEXT PRIMARY KEY,
    encrypted_email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL UNIQUE,
	encrypted_birthdate  TEXT NOT NULL,
    encrypted_lastname TEXT NOT NULL,
    encrypted_firstname TEXT NOT NULL,
	encrypted_gender TEXT NOT NULL,
	encrypted_telephone TEXT NOT NULL UNIQUE,
    encrypted_created_at TEXT,
    encrypted_postal_code TEXT NOT NULL, 
    encrypted_city TEXT NOT NULL, 
    encrypted_address1 TEXT NOT NULL, 
    encrypted_address2 TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE unverified_users;
-- +goose StatementEnd
