-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS votes (
    user_id TEXT NOT NULL,
    days TEXT NOT NULL CHECK(length(days) > 0),
    month INTEGER NOT NULL CHECK(month > 0 and month < 13),
	year  INTEGER NOT NULL CHECK(year >= 2024),
    PRIMARY KEY(user_id, month, year)
    FOREIGN KEY (user_id) REFERENCES users (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE votes;
-- +goose StatementEnd
