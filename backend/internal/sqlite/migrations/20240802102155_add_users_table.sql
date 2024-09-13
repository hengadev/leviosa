-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
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
	address TEXT NOT NULL,
	city TEXT NOT NULL,
	postalcard INTEGER NOT NULL,
    oauth_providers TEXT,
    oauth_ids TEXT,
    UNIQUE(lastname, firstname)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
