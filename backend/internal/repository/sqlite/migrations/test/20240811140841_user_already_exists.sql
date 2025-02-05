-- +goose Up
-- +goose StatementBegin
-- TODO: change that migration to use the right values for the columns
INSERT INTO users (
    email_hash,
    encrypted_email,
    password_hash,
    encrypted_picture,
    encrypted_created_at,
    encrypted_logged_in_at,
    role,
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
    'john.doe@example.com',
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
