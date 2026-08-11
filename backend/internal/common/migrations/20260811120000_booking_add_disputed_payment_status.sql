-- +goose Up
-- +goose StatementBegin

-- Add 'disputed' as a valid payment_status value (issue 018). A dispute is
-- recorded via a Stripe charge.dispute.created webhook and is just another
-- payment-status value on the existing booking-derived model (ADR-0007) —
-- no new table or Stripe-sourced reporting layer.
ALTER TABLE booking.bookings DROP CONSTRAINT bookings_payment_status_check;
ALTER TABLE booking.bookings ADD CONSTRAINT bookings_payment_status_check
    CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded', 'disputed'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE booking.bookings DROP CONSTRAINT bookings_payment_status_check;
ALTER TABLE booking.bookings ADD CONSTRAINT bookings_payment_status_check
    CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded'));

-- +goose StatementEnd
