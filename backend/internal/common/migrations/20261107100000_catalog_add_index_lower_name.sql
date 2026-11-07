-- +goose Up
-- +goose StatementBegin

-- Product and category names are kept as typed, for display, and are unique
-- whatever their case: "Ostéopathie" and "ostéopathie" are the same name.
-- The plain UNIQUE constraints on name stay for now (add first, remove
-- later): the previous release lowercases every name it writes, so it never
-- trips these indexes. Lookups by name compare lower(name), which they use.
CREATE UNIQUE INDEX catalog_categories_name_lower_key ON catalog.categories (lower(name));
CREATE UNIQUE INDEX catalog_products_name_lower_key ON catalog.products (lower(name));

COMMENT ON COLUMN catalog.categories.name IS 'Name of the category, as typed; unique ignoring case (catalog_categories_name_lower_key).';
COMMENT ON COLUMN catalog.products.name IS 'Name of the product, as typed; unique ignoring case (catalog_products_name_lower_key).';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS catalog.catalog_products_name_lower_key;
DROP INDEX IF EXISTS catalog.catalog_categories_name_lower_key;

COMMENT ON COLUMN catalog.categories.name IS 'Name of the category (must be unique).';
COMMENT ON COLUMN catalog.products.name IS 'Name of the product (must be unique).';

-- +goose StatementEnd
