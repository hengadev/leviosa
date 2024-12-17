-- +goose Up
-- +goose StatementBegin
INSERT INTO available_votes (days, month, year) VALUES
    ('23|12|3|9|17', 4, 2025),
    ('12|18|5', 5, 2025),
    ('7|21', 6, 2025),
    ('18|2|30', 7, 2025);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE from available_votes ORDER BY id DESC LIMIT 4;
-- +goose StatementEnd
