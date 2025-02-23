-- +goose Up
-- +goose StatementBegin
INSERT INTO products (
    id,
    name,
    description
    ) VALUES (
    'f1625792-6363-4111-943a-547d68d76d15',
    'First product name for Leviosa',
    'First product description for Leviosa'
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM products WHERE rowid = 1;
-- +goose StatementEnd
