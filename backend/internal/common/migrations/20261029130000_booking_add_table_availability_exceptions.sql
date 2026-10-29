-- +goose Up
-- +goose StatementBegin

-- Availability exceptions (issue 012): a date or a date range (civil dates in
-- Europe/Paris, both ends inclusive) on which a recurring availability does
-- not apply. No free-text column (ADR-0019: no health data).
--
-- Compatible with the previous release: a new table it neither reads nor
-- writes. Deleting an availability row (tests only; the app deactivates)
-- takes its exceptions with it.
CREATE TABLE booking.availability_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    availability_id UUID NOT NULL REFERENCES booking.availabilities(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT availability_exceptions_range_check CHECK (end_date >= start_date)
);

CREATE INDEX idx_availability_exceptions_availability_id
    ON booking.availability_exceptions (availability_id, start_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS booking.availability_exceptions;

-- +goose StatementEnd
