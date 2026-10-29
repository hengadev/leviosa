-- +goose Up
-- +goose StatementBegin

-- Client cancellation with a tiered refund (ADR-0014, issue 017).
--
-- Compatible with the previous release: every new column is nullable and the
-- payment status check only gains a value. The previous release neither
-- writes nor reads the new columns, and no existing row is rewritten.

-- What the client paid online for the booking (the deposit, or the full
-- price), set when the booking is confirmed: from the payment-succeeded
-- webhook, or 0 when nothing was due online. NULL on older bookings: the
-- refund then falls back to the product's current deposit applied to the
-- booking's total price.
ALTER TABLE booking.bookings ADD COLUMN amount_paid_online_cents INTEGER
    CHECK (amount_paid_online_cents IS NULL OR amount_paid_online_cents >= 0);

-- Who cancelled the booking. NULL on bookings not cancelled, and on those
-- cancelled before it was recorded (or by a payment cancellation).
ALTER TABLE booking.bookings ADD COLUMN cancelled_by VARCHAR(20)
    CHECK (cancelled_by IS NULL OR cancelled_by IN ('client', 'practitioner'));

-- What the cancellation refunded to the client, in cents (0 when the tier
-- refunds nothing). NULL when the booking was not cancelled this way.
ALTER TABLE booking.bookings ADD COLUMN refunded_amount_cents INTEGER
    CHECK (refunded_amount_cents IS NULL OR refunded_amount_cents >= 0);

-- A cancellation tier can refund part of the amount paid online.
ALTER TABLE booking.bookings DROP CONSTRAINT bookings_payment_status_check;
ALTER TABLE booking.bookings ADD CONSTRAINT bookings_payment_status_check
    CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded', 'partially_refunded', 'disputed'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE booking.bookings DROP CONSTRAINT bookings_payment_status_check;
ALTER TABLE booking.bookings ADD CONSTRAINT bookings_payment_status_check
    CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded', 'disputed'));
ALTER TABLE booking.bookings DROP COLUMN IF EXISTS refunded_amount_cents;
ALTER TABLE booking.bookings DROP COLUMN IF EXISTS cancelled_by;
ALTER TABLE booking.bookings DROP COLUMN IF EXISTS amount_paid_online_cents;

-- +goose StatementEnd
