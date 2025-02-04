-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT,
    picture TEXT,
	created_at  TEXT,
	logged_in_at TEXT,
    role TEXT NOT NULL,
	birthdate  TEXT NOT NULL,
    lastname TEXT NOT NULL,
    firstname TEXT NOT NULL,
	gender TEXT NOT NULL CHECK (GENDER IN ("M", "F", "NB", "NP")),
	telephone TEXT NOT NULL UNIQUE,
    postal_code TEXT NOT NULL, 
    city TEXT NOT NULL, 
    address1 TEXT NOT NULL, 
    address2 TEXT NOT NULL, 
    google_id TEXT NOT NULL, 
    apple_id TEXT NOT NULL, 
    UNIQUE(lastname, firstname)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
