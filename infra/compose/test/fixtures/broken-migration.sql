-- HARNESS FIXTURE — copied into a temporary copy of backend/ to build an
-- image whose `migrate` fails, proving a failed migration aborts the deploy
-- before anything is recreated. Never part of the real migrations.

-- +goose Up
SELECT harness_deliberately_missing_function();

-- +goose Down
SELECT 1;
