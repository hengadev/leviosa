-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS unverified_users (
    email TEXT PRIMARY KEY,
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
    address2 TEXT NOT NULL UNIQUE 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE unverified_users;
-- +goose StatementEnd
