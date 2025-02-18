-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS event_offers (
    event_id TEXT NOT NULL,
    offer_id TEXT NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events(id),
    FOREIGN KEY (offer_id) REFERENCES offers(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE event_offers;
-- +goose StatementEnd
