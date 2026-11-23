-- +goose Up
-- +goose StatementBegin

-- When the practitioner marked the booking as a no-show (issue 038).
-- NULL on bookings never marked as a no-show, and on those reopened the same
-- day. Nullable and additive: the previous release neither reads nor writes
-- it, and no existing row is rewritten.

ALTER TABLE booking.bookings ADD COLUMN no_show_at TIMESTAMPTZ;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE booking.bookings DROP COLUMN IF EXISTS no_show_at;

-- +goose StatementEnd
