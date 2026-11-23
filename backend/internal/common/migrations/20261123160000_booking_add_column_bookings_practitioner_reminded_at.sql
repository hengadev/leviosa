-- +goose Up
-- +goose StatementBegin

-- When the practitioner's "Dans 15 min" session notification was sent
-- (issue 046). NULL on bookings never reminded. Nullable and additive: the
-- previous release neither reads nor writes it, and no existing row is
-- rewritten.

ALTER TABLE booking.bookings ADD COLUMN practitioner_reminded_at TIMESTAMPTZ;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE booking.bookings DROP COLUMN IF EXISTS practitioner_reminded_at;

-- +goose StatementEnd
