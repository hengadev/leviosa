-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS events (
    id TEXT NOT NULL PRIMARY KEY,
    location TEXT NOT NULL,        
    placecount INTEGER NOT NULL,
    freeplace INTEGER CHECK(freeplace >= 0),
    beginat TEXT NOT NULL,
    sessionduration INTEGER NOT NULL,
    priceid TEXT NOT NULL UNIQUE,
    day INTEGER NOT NULL,
    month INTEGER NOT NULL,           
    year INTEGER NOT NULL,
    UNIQUE(day, month, year)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE events;
-- +goose StatementEnd
