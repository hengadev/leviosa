-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
INSERT INTO votes ( userid, days, month, year)
    VALUES ( 1, '23|12|6', 4, 2025);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DELETE from votes ORDER BY id DESC LIMIT 1;
-- +goose StatementEnd
