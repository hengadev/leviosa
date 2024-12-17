-- +goose Up
-- +goose StatementBegin
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
    oauth_providers,
    oauth_ids
    ) VALUES (
    'john.doe@gmail.com',
    '$a9rfNhA$N$A78#m',
    '',
    '',
    'basic',
    '1998-07-12',
    'DOE',
    'John',
    'M',
    '0123456789',
    NULL,
    NULL
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE from users ORDER BY id DESC LIMIT 1;
-- +goose StatementEnd
