-- +goose Up
-- +goose StatementBegin
INSERT INTO users (
    email,
    password,
    picture,
    created_at,
    logged_in_at,
    role,
    birthdate,
    lastname,
    firstname,
    gender,
    telephone,
    postal_code,
    city,
    address1,
    address2,
    google_id,
    apple_id
    ) VALUES (
    'john.doe@example.com',
    'hashedpassword',
    'picture',
    '2025-02-03',
    '2025-02-03',
    'basic',
    '1998-07-12',
    'DOE',
    'John',
    'M',
    '0123456789',
    '75000',
    'Paris',
    '01 Avenue Jean DUPONT',
    '',
    'google_id',
    'apple_id'
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE from users ORDER BY id DESC LIMIT 1;
-- +goose StatementEnd
