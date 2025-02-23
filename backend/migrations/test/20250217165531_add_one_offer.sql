-- +goose Up 
-- +goose StatementBegin
INSERT INTO offers (
    id,
    product_id,
    name,
    description,
    encrypted_picture,
    duration,
    price,
    encrypted_price_id
    ) VALUES (
    'f1625792-6363-4111-943a-547d68d76d1t',
    '366eb4a7-7853-4854-ab08-d12c53af7503',
    'First offer name for Leviosa',
    'First offer description for Leviosa',
    'picture',
    30,
    18,
    '07df7a28-f5f0-48ac-a11a-3d8b96d7760e'
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM offers WHERE rowid = 1;
-- +goose StatementEnd
