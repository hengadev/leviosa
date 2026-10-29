package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Availability exceptions (docs/issues/012-availability-exceptions.md).
//
// An exception closes a recurring availability on a date or a range of dates
// (holidays, a training day). Dates are civil dates in the practice time zone
// (Europe/Paris), both ends inclusive: a single date is a range whose start
// and end are the same day. An occurrence is closed when the date it starts on
// (in the practice time zone) falls in the range. Closing a date never cancels
// or moves the bookings already in it.
//
// An exception holds no free text (ADR-0019: no health data): only its dates.

// Availability exception errors.
var (
	ErrExceptionRangeOrder        = errors.New("end_date must not be before start_date")
	ErrExceptionInThePast         = errors.New("end_date must not be in the past")
	ErrExceptionOnOneOff          = errors.New("only a recurring availability can have exceptions")
	ErrExceptionOnCancelledSeries = errors.New("the recurring availability has been deleted")
	ErrInvalidExceptionDate       = errors.New("dates must be formatted YYYY-MM-DD")
)

// AvailabilityException is a date range on which a recurring availability
// does not apply. StartDate and EndDate are civil dates (midnight UTC, the
// representation civilDate uses), EndDate inclusive.
type AvailabilityException struct {
	ID             uuid.UUID
	AvailabilityID uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	CreatedAt      time.Time
}

// NewAvailabilityException validates and builds an exception on a recurring
// availability. today is the current civil date in the practice time zone: a
// range that has already ended is refused (it would close nothing).
// Ownership is the caller's concern (the application layer compares the
// availability's partner with the caller).
func NewAvailabilityException(availability *Availability, startDate, endDate, today time.Time) (*AvailabilityException, error) {
	if availability == nil || availability.ID == uuid.Nil {
		return nil, ErrInvalidAvailabilityID
	}
	if !availability.IsRecurring || availability.RecurrencePattern == nil {
		return nil, ErrExceptionOnOneOff
	}
	if availability.Status == AvailabilityStatusCancelled {
		return nil, ErrExceptionOnCancelledSeries
	}
	startDate, endDate, today = civilDate(startDate), civilDate(endDate), civilDate(today)
	if endDate.Before(startDate) {
		return nil, ErrExceptionRangeOrder
	}
	if endDate.Before(today) {
		return nil, ErrExceptionInThePast
	}
	return &AvailabilityException{
		ID:             uuid.New(),
		AvailabilityID: availability.ID,
		StartDate:      startDate,
		EndDate:        endDate,
		CreatedAt:      time.Now(),
	}, nil
}

// ParseCivilDate parses a YYYY-MM-DD date into the civil date representation
// exceptions use.
func ParseCivilDate(value string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, ErrInvalidExceptionDate
	}
	return date, nil
}

// PracticeToday is the current civil date in the practice time zone.
func PracticeToday(now time.Time) time.Time {
	return civilDate(now.In(PracticeLocation))
}

// CoversDate reports whether the civil date day falls in the exception.
func (e AvailabilityException) CoversDate(day time.Time) bool {
	day = civilDate(day)
	return !day.Before(e.StartDate) && !day.After(e.EndDate)
}

// CoversInstant reports whether the instant's date, in the practice time
// zone, falls in the exception.
func (e AvailabilityException) CoversInstant(instant time.Time) bool {
	return e.CoversDate(civilDate(instant.In(PracticeLocation)))
}
