-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS votes (
    userid INTEGER NOT NULL CHECK(userid > 0),
    days TEXT NOT NULL CHECK(length(days) > 0),
    month INTEGER NOT NULL CHECK(month > 0 and month < 13),
	year  INTEGER NOT NULL CHECK(year >= 2024),
    PRIMARY KEY(userid, month, year)
    FOREIGN KEY (userid) REFERENCES users (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE votes;
-- +goose StatementEnd
