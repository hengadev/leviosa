-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT,
	createdat  DATETIME,
	loggedinat DATETIME,
    role TEXT NOT NULL,
	birthdate  TEXT NOT NULL,
    lastname TEXT NOT NULL,
    firstname TEXT NOT NULL,
	gender TEXT NOT NULL CHECK (GENDER IN ("M", "F", "NB", "NP")),
	telephone TEXT NOT NULL UNIQUE,
    postal_code TEXT NOT NULL UNIQUE, 
    city TEXT NOT NULL UNIQUE, 
    address1 TEXT NOT NULL UNIQUE, 
    address2 TEXT NOT NULL UNIQUE,
    google_id TEXT,
    apple_id TEXT,
    UNIQUE(lastname, firstname)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
