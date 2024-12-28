-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pending_users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT,
	birthdate  TEXT NOT NULL,
    lastname TEXT NOT NULL,
    firstname TEXT NOT NULL,
	gender TEXT NOT NULL,
	telephone TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    postal_code TEXT NOT NULL UNIQUE, 
    city TEXT NOT NULL UNIQUE, 
    address1 TEXT NOT NULL UNIQUE, 
    address2 TEXT NOT NULL UNIQUE,
    google_id TEXT UNIQUE,
    apple_id TEXT UNIQUE 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE pending_users;
-- +goose StatementEnd
