package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Availability exceptions (issue 012): given a series, its exceptions and a
// range, which occurrences come out.

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func exceptionOn(a *Availability, start, end time.Time) AvailabilityException {
	return AvailabilityException{ID: uuid.New(), AvailabilityID: a.ID, StartDate: start, EndDate: end}
}

func TestOccurrencesExceptSingleDate(t *testing.T) {
	// Every day 09:00–13:00 from Monday 2026-11-02; Wednesday 4th closed.
	a := series(paris(2026, 11, 2, 9, 0), paris(2026, 11, 2, 13, 0), RecurrencePattern{Type: RecurrenceDaily, Interval: 1})
	exceptions := []AvailabilityException{exceptionOn(a, date(2026, 11, 4), date(2026, 11, 4))}

	windows := a.OccurrencesExcept(paris(2026, 11, 2, 0, 0), paris(2026, 11, 7, 0, 0), exceptions)

	assertInstants(t, []time.Time{
		paris(2026, 11, 2, 9, 0), paris(2026, 11, 3, 9, 0), paris(2026, 11, 5, 9, 0), paris(2026, 11, 6, 9, 0),
	}, occurrenceStarts(windows))
}

func TestOccurrencesExceptRange(t *testing.T) {
	// Every Monday and Thursday 09:00–13:00 from Monday 2026-12-14; closed
	// from Thursday 24 to Monday 28 December, both ends inclusive.
	a := series(paris(2026, 12, 14, 9, 0), paris(2026, 12, 14, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{1, 4},
	})
	exceptions := []AvailabilityException{exceptionOn(a, date(2026, 12, 24), date(2026, 12, 28))}

	windows := a.OccurrencesExcept(paris(2026, 12, 14, 0, 0), paris(2027, 1, 5, 0, 0), exceptions)

	assertInstants(t, []time.Time{
		paris(2026, 12, 14, 9, 0), paris(2026, 12, 17, 9, 0), paris(2026, 12, 21, 9, 0),
		// 24 and 28 December closed
		paris(2026, 12, 31, 9, 0), paris(2027, 1, 4, 9, 0),
	}, occurrenceStarts(windows))
}

func TestOccurrencesExceptUsesThePracticeDate(t *testing.T) {
	// A series starting at 00:30: on 5 November it starts at 23:30 UTC on
	// the 4th. The exception follows the Paris date, not the UTC one.
	a := series(paris(2026, 11, 2, 0, 30), paris(2026, 11, 2, 2, 0), RecurrencePattern{Type: RecurrenceDaily, Interval: 1})
	exceptions := []AvailabilityException{exceptionOn(a, date(2026, 11, 5), date(2026, 11, 5))}

	windows := a.OccurrencesExcept(paris(2026, 11, 4, 0, 0), paris(2026, 11, 7, 0, 0), exceptions)

	assertInstants(t, []time.Time{paris(2026, 11, 4, 0, 30), paris(2026, 11, 6, 0, 30)}, occurrenceStarts(windows))
}

func TestOccurrencesExceptIgnoresOtherAvailabilities(t *testing.T) {
	a := series(paris(2026, 11, 2, 9, 0), paris(2026, 11, 2, 13, 0), RecurrencePattern{Type: RecurrenceDaily, Interval: 1})
	other := series(paris(2026, 11, 2, 14, 0), paris(2026, 11, 2, 18, 0), RecurrencePattern{Type: RecurrenceDaily, Interval: 1})
	exceptions := []AvailabilityException{exceptionOn(other, date(2026, 11, 3), date(2026, 11, 3))}

	windows := a.OccurrencesExcept(paris(2026, 11, 2, 0, 0), paris(2026, 11, 4, 0, 0), exceptions)

	assertInstants(t, []time.Time{paris(2026, 11, 2, 9, 0), paris(2026, 11, 3, 9, 0)}, occurrenceStarts(windows))
}

func TestNewAvailabilityException(t *testing.T) {
	weekly := series(paris(2026, 11, 2, 9, 0), paris(2026, 11, 2, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{1},
	})
	today := date(2026, 11, 1)

	t.Run("accepts a single date", func(t *testing.T) {
		e, err := NewAvailabilityException(weekly, date(2026, 11, 9), date(2026, 11, 9), today)
		require.NoError(t, err)
		assert.Equal(t, weekly.ID, e.AvailabilityID)
		assert.True(t, e.CoversDate(date(2026, 11, 9)))
		assert.False(t, e.CoversDate(date(2026, 11, 10)))
	})

	t.Run("accepts a range ending today", func(t *testing.T) {
		_, err := NewAvailabilityException(weekly, date(2026, 10, 30), today, today)
		require.NoError(t, err)
	})

	t.Run("refuses a range ending before it starts", func(t *testing.T) {
		_, err := NewAvailabilityException(weekly, date(2026, 11, 10), date(2026, 11, 9), today)
		assert.ErrorIs(t, err, ErrExceptionRangeOrder)
	})

	t.Run("refuses a range already over", func(t *testing.T) {
		_, err := NewAvailabilityException(weekly, date(2026, 10, 20), date(2026, 10, 31), today)
		assert.ErrorIs(t, err, ErrExceptionInThePast)
	})

	t.Run("refuses a one-off availability", func(t *testing.T) {
		oneOff := &Availability{ID: uuid.New(), StartTime: paris(2026, 11, 2, 9, 0), EndTime: paris(2026, 11, 2, 13, 0)}
		_, err := NewAvailabilityException(oneOff, date(2026, 11, 2), date(2026, 11, 2), today)
		assert.ErrorIs(t, err, ErrExceptionOnOneOff)
	})

	t.Run("refuses a deleted series", func(t *testing.T) {
		deleted := *weekly
		deleted.Status = AvailabilityStatusCancelled
		_, err := NewAvailabilityException(&deleted, date(2026, 11, 9), date(2026, 11, 9), today)
		assert.ErrorIs(t, err, ErrExceptionOnCancelledSeries)
	})
}

func TestParseCivilDate(t *testing.T) {
	d, err := ParseCivilDate("2026-12-24")
	require.NoError(t, err)
	assert.Equal(t, date(2026, 12, 24), d)

	_, err = ParseCivilDate("24/12/2026")
	assert.ErrorIs(t, err, ErrInvalidExceptionDate)
}
