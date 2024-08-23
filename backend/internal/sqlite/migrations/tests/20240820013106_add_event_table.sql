-- +goose Up
-- +goose StatementBegin
-- TODO: add the check year >= 2024 on production database
CREATE TABLE IF NOT EXISTS events (
    id TEXT NOT NULL PRIMARY KEY CHECK(length(id) > 0),
    location TEXT NOT NULL CHECK(length(location) > 0),
    placecount INTEGER NOT NULL CHECK(placecount > 0),
    freeplace INTEGER CHECK(freeplace >= 0),
    beginat TEXT NOT NULL CHECK (beginat LIKE '__:__:__'),
    sessionduration INTEGER NOT NULL CHECK(sessionduration > 0),
    priceid TEXT NOT NULL UNIQUE CHECK(length(priceid) > 0),
    day INTEGER NOT NULL CHECK(day > 0),
    month INTEGER NOT NULL CHECK(month > 0),
    year INTEGER NOT NULL,
    UNIQUE(day, month, year)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE events;
-- +goose StatementEnd
