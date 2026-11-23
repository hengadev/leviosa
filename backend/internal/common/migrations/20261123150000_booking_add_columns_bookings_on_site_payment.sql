-- +goose Up
-- +goose StatementBegin

-- The on-site payment (issue 039): what the client paid at the practice, on
-- the card terminal, as the practitioner recorded it, and when it was
-- recorded. Both NULL until recorded, and again when undone the same day.
-- Nullable and additive: the previous release neither reads nor writes them,
-- and no existing row is rewritten.

ALTER TABLE booking.bookings ADD COLUMN amount_paid_on_site_cents INTEGER;
ALTER TABLE booking.bookings ADD COLUMN paid_on_site_at TIMESTAMPTZ;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE booking.bookings DROP COLUMN IF EXISTS paid_on_site_at;
ALTER TABLE booking.bookings DROP COLUMN IF EXISTS amount_paid_on_site_cents;

-- +goose StatementEnd
