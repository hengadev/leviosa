-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS event_products (
    event_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE event_products;
-- +goose StatementEnd
