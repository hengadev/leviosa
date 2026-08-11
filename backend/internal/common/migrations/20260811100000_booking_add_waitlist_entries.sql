-- +goose Up
-- +goose StatementBegin

-- Waitlist entries: client demand for a fully booked product/date range (PRD:
-- booking-lifecycle-policy-engine, issue 014). Deliberately a separate table
-- from booking.bookings — a waitlist entry has no slot, room allocation, or
-- payment yet, so it does not belong in the Booking FSM (ADR-0006).
CREATE TABLE booking.waitlist_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID, -- References auth.users.id (from authuser service); NULL for guest entries

    -- Desired product and date range (encrypted for GDPR compliance, same
    -- rationale as booking.bookings.slot_start_time/slot_end_time)
    product_id_encrypted BYTEA NOT NULL, -- References catalog.products.id
    desired_range_start_encrypted BYTEA NOT NULL,
    desired_range_end_encrypted BYTEA NOT NULL,

    -- Optional free-form notes from the client (encrypted)
    notes_encrypted BYTEA,

    -- Guest contact fields (encrypted, used when client_id IS NULL) — mirrors
    -- booking.bookings' guest columns (ADR-0008)
    guest_first_name_encrypted BYTEA,
    guest_last_name_encrypted BYTEA,
    guest_email_encrypted BYTEA,
    guest_phone_encrypted BYTEA,

    -- Lifecycle: waiting -> notified -> claimed | expired
    --            waiting | notified -> cancelled
    status VARCHAR(20) NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'notified', 'claimed', 'expired', 'cancelled')),
    notified_at TIMESTAMP WITH TIME ZONE,
    claimed_at TIMESTAMP WITH TIME ZONE,
    expired_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,

    -- Encryption metadata (required by encx)
    dek_encrypted BYTEA NOT NULL,
    key_version INTEGER NOT NULL DEFAULT 1,

    -- Administrative fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    metadata JSONB
);

CREATE INDEX idx_waitlist_entries_client ON booking.waitlist_entries(client_id, created_at DESC);
CREATE INDEX idx_waitlist_entries_status ON booking.waitlist_entries(status, created_at ASC);

CREATE TRIGGER waitlist_entries_update_timestamp BEFORE UPDATE ON booking.waitlist_entries
    FOR EACH ROW EXECUTE FUNCTION booking.update_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS booking.waitlist_entries;

-- +goose StatementEnd
