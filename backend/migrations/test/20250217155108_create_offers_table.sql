-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS offers (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL UNIQUE,
    encrypted_picture TEXT NOT NULL UNIQUE,
    duration INTEGER NOT NULL,
    price INTEGER NOT NULL,
    encrypted_price_id TEXT NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE offers;
-- +goose StatementEnd
