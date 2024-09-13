-- +goose Up
-- +goose StatementBegin
INSERT INTO users ( email, password, createdat, loggedinat, role, birthdate, lastname, firstname, gender, telephone, oauth_providers, oauth_ids)
VALUES
    ('jane.doe@gmail.com', 'w4w3f09QF&h)#fwe', '', '', 'basic', '1998-07-12', 'DOE', 'Jane', 'F', '0123456780', NULL, NULL),
    ('jean.doe@gmail.com', 'wf0fT^9f2$$_aewa', '', '', 'basic', '1998-07-12', 'DOE', 'Jean', 'M', '0123456781', NULL, NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE from users ORDER BY id DESC LIMIT 3;
-- +goose StatementEnd
