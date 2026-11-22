-- +goose Up
-- +goose StatementBegin

-- The verification code duration (otp_duration) is in seconds: the admin
-- settings page and its endpoint accept 60 to 3600. The initial schema seeded
-- it as '15', meant as minutes, which read as seconds is out of range. Set it
-- to 600 (10 minutes, the duration codes have had so far). Only the untouched
-- seed changes: the endpoint never accepts 15, so a value an administrator
-- saved is left as is. The previous release ignores this setting, so the
-- change is safe for it.
UPDATE settings.plain SET value = '600' WHERE key = 'otp_duration' AND value = '15';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Nothing to undo: '600' is valid for every release.
SELECT 1;

-- +goose StatementEnd
