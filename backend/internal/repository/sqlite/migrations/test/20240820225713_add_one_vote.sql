-- +goose Up
-- +goose StatementBegin
INSERT INTO votes ( userid, days, month, year)
    VALUES ( 1, '23|12|6', 4, 2025);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM votes WHERE rowid = (SELECT MAX(rowid) FROM votes);
-- +goose StatementEnd
