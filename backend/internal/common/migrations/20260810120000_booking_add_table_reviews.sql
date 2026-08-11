-- +goose Up
-- +goose StatementBegin

-- Reviews: a client's rating + optional comment for a completed booking (issue 009).
-- One review per booking (enforced by the unique constraint on booking_id), tied to the
-- booking's client/partner/product so public reads can be filtered without a join back
-- to booking.bookings (whose client/product fields are encrypted).
--
-- Unlike most booking.* tables, review text is NOT encx-encrypted: reviews exist to be
-- displayed publicly (see docs/prd/public-trust-and-seo-foundations.md), so there is no
-- GDPR-at-rest requirement pushing toward encryption the way there is for client notes
-- or contact details.
CREATE TABLE booking.reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL UNIQUE REFERENCES booking.bookings(id) ON DELETE CASCADE,
    client_id UUID NOT NULL,
    partner_id UUID NOT NULL,
    product_id UUID NOT NULL,

    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,

    -- Moderation: hidden (not deleted) reviews are excluded from public reads but
    -- retained for an audit trail (see ADR precedent for hide-not-delete moderation).
    hidden BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Public read pattern: reviews for a partner/product, excluding hidden ones.
CREATE INDEX idx_reviews_partner_visible ON booking.reviews(partner_id, created_at DESC) WHERE hidden = FALSE;
CREATE INDEX idx_reviews_product_visible ON booking.reviews(product_id, created_at DESC) WHERE hidden = FALSE;

-- Moderation view: all reviews for a partner/product regardless of visibility.
CREATE INDEX idx_reviews_partner_all ON booking.reviews(partner_id, created_at DESC);
CREATE INDEX idx_reviews_product_all ON booking.reviews(product_id, created_at DESC);

CREATE TRIGGER reviews_update_timestamp BEFORE UPDATE ON booking.reviews
    FOR EACH ROW EXECUTE FUNCTION booking.update_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS booking.reviews;

-- +goose StatementEnd
