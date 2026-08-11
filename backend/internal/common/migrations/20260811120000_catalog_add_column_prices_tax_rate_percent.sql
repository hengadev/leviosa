-- +goose Up
-- +goose StatementBegin
ALTER TABLE catalog.prices
    ADD COLUMN tax_rate_percent NUMERIC(5,2) DEFAULT NULL;

ALTER TABLE catalog.prices
    ADD CONSTRAINT chk_prices_tax_rate_percent_range
    CHECK (tax_rate_percent IS NULL OR (tax_rate_percent >= 0 AND tax_rate_percent <= 100));

COMMENT ON COLUMN catalog.prices.tax_rate_percent IS
    'Optional VAT/tax rate as a percentage (e.g. 20.00 for 20%). NULL means exempt/unset — the safe default for a business that may operate under the French franchise en base de TVA regime. Stored and applied at display/calculation time; not synced to Stripe Tax.';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE catalog.prices DROP CONSTRAINT IF EXISTS chk_prices_tax_rate_percent_range;
ALTER TABLE catalog.prices DROP COLUMN IF EXISTS tax_rate_percent;
-- +goose StatementEnd
