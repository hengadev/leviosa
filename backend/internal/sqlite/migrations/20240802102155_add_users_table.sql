-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
	createdat  DATETIME,
	loggedinat DATETIME,
    role TEXT NOT NULL,
	birthdate  DATETIME,
    lastname TEXT NOT NULL,
    firstname TEXT NOT NULL,
	gender TEXT NOT NULL CHECK (GENDER IN ("M", "F", "NB", "NP")),
	telephone TEXT NOT NULL UNIQUE,
	address TEXT NOT NULL,
	city TEXT NOT NULL,
	postalcard INTEGER NOT NULL,
    UNIQUE(lastname, firstname)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE users;
-- +goose StatementEnd
