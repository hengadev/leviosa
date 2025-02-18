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
    ) VALUES
        (
            '42002e73-cfd4-4e2b-a914-06ddee24823a',
            '67c64adb-cf07-423d-9500-9211a139dacf',
            'Second offer name for Leviosa',
            'Second offer description for Leviosa',
            'picture2',
            30,
            20,
            '1fe9daa1-b6c9-4c1f-8ec2-a3c8274c6211'
        ),
        (
            '472f99f7-b15f-4dee-8e49-2fc28c8cd28e',
            '5983ef96-5b98-4267-ba44-4abb6ad8f6db',
            'Third offer name for Leviosa',
            'Third offer description for Leviosa',
            'picture3',
            10,
            50,
            'f148df20-b6c4-4e9e-9f3d-5b88ac80c4f7'
        );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM offers WHERE rowid > 2;
-- +goose StatementEnd
