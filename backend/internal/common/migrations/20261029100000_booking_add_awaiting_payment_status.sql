-- +goose Up
-- +goose StatementBegin

-- Bookings await payment before confirmation (ADR-0018, issue 014): a new
-- booking starts 'awaiting_payment' and holds its slot until hold_expires_at,
-- then becomes 'confirmed' (payment succeeded) or 'expired' (hold elapsed).
--
-- Compatible with the previous release: the status check only gains values,
-- and the new columns are nullable — the previous release neither writes nor
-- reads them, and no existing row is rewritten.
ALTER TABLE booking.bookings DROP CONSTRAINT bookings_status_check;
ALTER TABLE booking.bookings ADD CONSTRAINT bookings_status_check
    CHECK (status IN ('awaiting_payment', 'expired', 'confirmed', 'cancelled', 'completed', 'no_show'));

-- End of the payment hold; set while the booking awaits payment.
ALTER TABLE booking.bookings ADD COLUMN hold_expires_at TIMESTAMP WITH TIME ZONE;

-- Language the client booked in, for notifications. NULL on older bookings
-- (read as French).
ALTER TABLE booking.bookings ADD COLUMN locale VARCHAR(2)
    CHECK (locale IS NULL OR locale IN ('fr', 'en'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE booking.bookings DROP COLUMN IF EXISTS locale;
ALTER TABLE booking.bookings DROP COLUMN IF EXISTS hold_expires_at;
ALTER TABLE booking.bookings DROP CONSTRAINT bookings_status_check;
ALTER TABLE booking.bookings ADD CONSTRAINT bookings_status_check
    CHECK (status IN ('confirmed', 'cancelled', 'completed', 'no_show'));

-- +goose StatementEnd
