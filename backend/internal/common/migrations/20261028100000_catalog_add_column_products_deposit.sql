-- +goose Up
-- +goose StatementBegin

-- Add the deposit JSONB column (issue 013, ADR-0018): the part of a
-- product's price paid online when booking, either a fixed amount in cents
-- or a percentage of the price. Nullable: a product with no deposit keeps
-- behaving exactly as before (the full price is paid online).
ALTER TABLE catalog.products ADD COLUMN deposit JSONB;

COMMENT ON COLUMN catalog.products.deposit IS 'Optional deposit paid online at booking: {type: "fixed"|"percent", value}. fixed value is cents, percent value is whole percentage points of the price. NULL means the full price is paid online.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE catalog.products DROP COLUMN IF EXISTS deposit;

-- +goose StatementEnd
