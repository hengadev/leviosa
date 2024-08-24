-- +goose Up
-- +goose StatementBegin
-- TODO:
-- - delete everything from the users table, insert with parameters
-- - create the use with the hashedpassword
DELETE FROM users;

-- +goose ENVSUB ON
INSERT INTO users (
    email,
    password,
    createdat,
    loggedinat,
    role,
    birthdate,
    lastname,
    firstname,
    gender,
    telephone,
    address,
    city,
    postalcard
    ) VALUES (
    'john.doe@gmail.com',
    '${HASHED_PASSWORD}',
    '',
    '',
    'basic',
    '1998-07-12',
    'DOE',
    'John',
    'M',
    '0123456789',
    'Impasse Inconnue',
    'Paris',
    12345
    );
-- +goose ENVSUB OFF
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE from users ORDER BY id DESC LIMIT 1;
-- +goose StatementEnd
