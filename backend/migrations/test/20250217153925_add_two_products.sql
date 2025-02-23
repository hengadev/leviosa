-- +goose Up
-- +goose StatementBegin
INSERT INTO products (
    id,
    name,
    description
    ) VALUES
        (
            '893a7ff5-bc34-438a-a0ed-1d426711e77a',
            'Second product name for Leviosa',
            'Second product description for Leviosa'
        ),
        (
            '019575bc-8494-45d9-9ca2-85bafa86a64f',
            'Third product name for Leviosa',
            'Third product description for Leviosa'
        );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM products WHERE rowid > 2;
-- +goose StatementEnd
