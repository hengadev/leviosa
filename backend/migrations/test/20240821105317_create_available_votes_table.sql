-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS available_votes (
    days TEXT NOT NULL CHECK(length(days) > 0),
    month INTEGER NOT NULL CHECK(month > 0 and month < 13),
	year  INTEGER NOT NULL CHECK(year >= 2024),
    PRIMARY KEY(month, year)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE available_votes;
-- +goose StatementEnd
