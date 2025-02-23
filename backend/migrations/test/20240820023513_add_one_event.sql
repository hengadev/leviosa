-- +goose Up 
-- +goose StatementBegin
INSERT INTO events (
    id,
    encrypted_title,
    encrypted_description,
    encrypted_city,
    encrypted_postal_code,
    encrypted_address1,
    encrypted_address2,
    placecount,
    freeplace,
    encrypted_begin_at,
    encrypted_end_at,
    encrypted_price_id,
    day,
    month,
    year
    ) VALUES (
    'ea1d74e2-1612-47ec-aee9-c6a46b65640f',
    'First event for Leviosa',
    'First description for Leviosa',
    'Paris',
    'postalCode',
    'address1',
    '',
    16,
    14,
    '08:00:00',
    '20:00:00',
    '179cf8f1-81ad-4ec1-b8bb-8f48abf9ef80',
    22,
    4,
    2025
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM events WHERE rowid = 1;
-- +goose StatementEnd
