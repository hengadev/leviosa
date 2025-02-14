-- +goose Up
-- +goose StatementBegin
INSERT INTO events (
    id,
    location,
    placecount,
    freeplace,
    beginat,
    sessionduration,
    priceid,
    day,
    month,
    year
    ) VALUES (
    'ea1d74e2-1612-47ec-aee9-c6a46b65640f',
    'Impasse Inconnue',
    16,
    14,
    '08:00:00',
    30,
    '4fe0vuw3ef0223',
    12,
    7,
    1998
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM events WHERE rowid = 1;
-- +goose StatementEnd
