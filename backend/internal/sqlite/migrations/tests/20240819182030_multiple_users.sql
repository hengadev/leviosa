-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
INSERT INTO users ( email, password, createdat, loggedinat, role, birthdate, lastname, firstname, gender, telephone, address, city, postalcard)
VALUES
    ('jane.doe@gmail.com', 'w4w3f09QF&h)#fwe', '', '', 'basic', '1998-07-12', 'DOE', 'Jane', 'F', '0123456780', 'Impasse Inconnue', 'Paris', 12345),
    ('jean.doe@gmail.com', 'wf0fT^9f2$$_aewa', '', '', 'basic', '1998-07-12', 'DOE', 'Jean', 'M', '0123456781', 'Impasse Inconnue', 'Paris', 12345);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DELETE from users ORDER BY id DESC LIMIT 3;
-- +goose StatementEnd
