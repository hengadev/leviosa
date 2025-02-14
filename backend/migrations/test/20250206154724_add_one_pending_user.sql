-- +goose Up
-- +goose StatementBegin
INSERT INTO pending_users (
    id,
    email_hash,
    encrypted_email,
    password_hash,
    encrypted_picture,
    encrypted_created_at,
    encrypted_birthdate,
    encrypted_lastname,
    encrypted_firstname,
    encrypted_gender,
    encrypted_telephone,
    encrypted_postal_code,
    encrypted_city,
    encrypted_address1,
    encrypted_address2,
    encrypted_google_id,
    encrypted_apple_id
) VALUES (
    "123e4567-e89b-12d3-a456-426614174000",
    'john.doe@example.com',
    'john.doe@example.com',
    'hashedpassword',
    'picture',
    '2025-02-03',
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
DELETE FROM pending_users WHERE rowid = 1;
-- +goose StatementEnd
