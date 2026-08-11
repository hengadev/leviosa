-- +goose Up
-- +goose StatementBegin

-- Add cancellation_tiers JSONB column for the graduated cancellation-fee
-- schedule (issue 016, ADR-0014 superseding ADR-0006's binary
-- cancellation-window rule). Nullable: a product with no tiers configured
-- keeps behaving exactly as before (governed by cancellation_hours alone)
-- until issue 017 wires tier-based refund calculation into CancelBooking.
ALTER TABLE catalog.products ADD COLUMN cancellation_tiers JSONB;

COMMENT ON COLUMN catalog.products.cancellation_tiers IS 'Ordered list of {minHoursBefore, refundPercentage} cancellation fee tiers. NULL/empty means the product has not opted into graduated fees yet.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE catalog.products DROP COLUMN IF EXISTS cancellation_tiers;

-- +goose StatementEnd
