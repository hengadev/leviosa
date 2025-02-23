-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events (
    id TEXT NOT NULL PRIMARY KEY,
    encrypted_title TEXT NOT NULL,
    encrypted_description TEXT NOT NULL,
    encrypted_postal_code TEXT NOT NULL,
    encrypted_city TEXT NOT NULL,
    encrypted_address1 TEXT NOT NULL,
    encrypted_address2 TEXT,
    placecount INTEGER NOT NULL,
    freeplace INTEGER CHECK(freeplace >= 0),
    encrypted_begin_at TEXT NOT NULL,
    encrypted_end_at TEXT NOT NULL,
    encrypted_price_id TEXT NOT NULL UNIQUE,
    day INTEGER NOT NULL CHECK(day > 0 AND day < 31) ,
    month INTEGER NOT NULL CHECK(month > 0 AND month < 13) ,
    year INTEGER NOT NULL,
    UNIQUE(day, month, year)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE events;
-- +goose StatementEnd
