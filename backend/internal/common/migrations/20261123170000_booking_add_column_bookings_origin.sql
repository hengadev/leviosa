-- +goose Up
-- +goose StatementBegin

-- How the booking was made (issue 041, ADR-0021): 'online' (the practice
-- page; every existing booking) or 'practitioner' (booked by the
-- practitioner for a client who called). Additive and compatible with the
-- previous release: it never writes the column, and its inserts take the
-- default.

ALTER TABLE booking.bookings
    ADD COLUMN origin TEXT NOT NULL DEFAULT 'online'
    CONSTRAINT bookings_origin_check CHECK (origin IN ('online', 'practitioner'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE booking.bookings DROP COLUMN IF EXISTS origin;

-- +goose StatementEnd
