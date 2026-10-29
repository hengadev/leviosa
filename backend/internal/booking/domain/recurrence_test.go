package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Recurrence expansion (issue 011): given a series and a range, which
// occurrence windows come out. Times are wall-clock times in the practice
// time zone, as the partner enters them.

func paris(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, PracticeLocation)
}

func series(start, end time.Time, pattern RecurrencePattern) *Availability {
	return &Availability{
		ID:                uuid.New(),
		StartTime:         start,
		EndTime:           end,
		IsRecurring:       true,
		RecurrencePattern: &pattern,
		Status:            AvailabilityStatusAvailable,
	}
}

func occurrenceStarts(windows []AvailabilityWindow) []time.Time {
	out := make([]time.Time, 0, len(windows))
	for _, w := range windows {
		out = append(out, w.Start)
	}
	return out
}

func assertInstants(t *testing.T, expected, actual []time.Time) {
	t.Helper()
	require.Len(t, actual, len(expected), "occurrences: %v", actual)
	for i := range expected {
		assert.True(t, expected[i].Equal(actual[i]), "occurrence %d: expected %s, got %s", i, expected[i], actual[i])
	}
}

func TestOccurrencesWeekly(t *testing.T) {
	// Every Monday 09:00–13:00 from Monday 2026-11-02.
	a := series(paris(2026, 11, 2, 9, 0), paris(2026, 11, 2, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{1},
	})

	windows := a.Occurrences(paris(2026, 11, 1, 0, 0), paris(2026, 11, 30, 0, 0))

	assertInstants(t, []time.Time{
		paris(2026, 11, 2, 9, 0), paris(2026, 11, 9, 9, 0), paris(2026, 11, 16, 9, 0), paris(2026, 11, 23, 9, 0),
	}, occurrenceStarts(windows))
	for _, w := range windows {
		assert.Equal(t, a.ID, w.AvailabilityID)
		assert.Equal(t, 4*time.Hour, w.End.Sub(w.Start))
	}
}

func TestOccurrencesWeeklyOnSeveralDays(t *testing.T) {
	// Mondays and Thursdays from Thursday 2026-11-05: the anchor's Monday
	// (11-02) is before the series starts.
	a := series(paris(2026, 11, 5, 14, 0), paris(2026, 11, 5, 18, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{1, 4},
	})

	windows := a.Occurrences(paris(2026, 11, 1, 0, 0), paris(2026, 11, 13, 0, 0))

	assertInstants(t, []time.Time{
		paris(2026, 11, 5, 14, 0), paris(2026, 11, 9, 14, 0), paris(2026, 11, 12, 14, 0),
	}, occurrenceStarts(windows))
}

func TestOccurrencesSeriesStartedBeforeTheRange(t *testing.T) {
	// A series anchored months ago still yields its occurrences in the range.
	a := series(paris(2026, 1, 5, 9, 0), paris(2026, 1, 5, 12, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{1},
	})

	windows := a.Occurrences(paris(2026, 11, 9, 0, 0), paris(2026, 11, 10, 0, 0))

	assertInstants(t, []time.Time{paris(2026, 11, 9, 9, 0)}, occurrenceStarts(windows))
}

func TestOccurrencesWeeklyInterval(t *testing.T) {
	// Every other week on Wednesday, from Wednesday 2026-11-04.
	a := series(paris(2026, 11, 4, 9, 0), paris(2026, 11, 4, 12, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 2, DaysOfWeek: []int{3},
	})

	windows := a.Occurrences(paris(2026, 11, 1, 0, 0), paris(2026, 12, 31, 0, 0))

	assertInstants(t, []time.Time{
		paris(2026, 11, 4, 9, 0), paris(2026, 11, 18, 9, 0), paris(2026, 12, 2, 9, 0),
		paris(2026, 12, 16, 9, 0), paris(2026, 12, 30, 9, 0),
	}, occurrenceStarts(windows))
}

func TestOccurrencesWeeklyIntervalCountsWeeksFromMonday(t *testing.T) {
	// Every other week on Monday and Saturday, anchored on Saturday
	// 2026-11-07: the anchor's week (Mon 11-02 – Sun 11-08) is week 0, so the
	// next occurrences are Monday 11-16 and Saturday 11-21.
	a := series(paris(2026, 11, 7, 9, 0), paris(2026, 11, 7, 12, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 2, DaysOfWeek: []int{1, 6},
	})

	windows := a.Occurrences(paris(2026, 11, 1, 0, 0), paris(2026, 11, 23, 0, 0))

	assertInstants(t, []time.Time{
		paris(2026, 11, 7, 9, 0), paris(2026, 11, 16, 9, 0), paris(2026, 11, 21, 9, 0),
	}, occurrenceStarts(windows))
}

func TestOccurrencesDailyInterval(t *testing.T) {
	// Every three days from 2026-11-02.
	a := series(paris(2026, 11, 2, 10, 0), paris(2026, 11, 2, 11, 0), RecurrencePattern{
		Type: RecurrenceDaily, Interval: 3,
	})

	windows := a.Occurrences(paris(2026, 11, 3, 0, 0), paris(2026, 11, 12, 0, 0))

	assertInstants(t, []time.Time{
		paris(2026, 11, 5, 10, 0), paris(2026, 11, 8, 10, 0), paris(2026, 11, 11, 10, 0),
	}, occurrenceStarts(windows))
}

func TestOccurrencesEndDate(t *testing.T) {
	// Weekly on Monday until Monday 2026-11-16, inclusive: Until holds a time
	// earlier than the occurrence's start, and the day still counts.
	until := paris(2026, 11, 16, 0, 0)
	a := series(paris(2026, 11, 2, 9, 0), paris(2026, 11, 2, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{1}, Until: &until,
	})

	windows := a.Occurrences(paris(2026, 11, 1, 0, 0), paris(2027, 1, 1, 0, 0))

	assertInstants(t, []time.Time{
		paris(2026, 11, 2, 9, 0), paris(2026, 11, 9, 9, 0), paris(2026, 11, 16, 9, 0),
	}, occurrenceStarts(windows))

	assert.False(t, a.SeriesEndedBefore(paris(2026, 11, 16, 20, 0)), "the last day is still part of the series")
	assert.True(t, a.SeriesEndedBefore(paris(2026, 11, 17, 0, 0)))
}

func TestOccurrencesKeepWallClockAcrossDaylightSaving(t *testing.T) {
	// Every Monday 09:00–13:00 Paris across the spring change (Sunday
	// 2027-03-28) and the autumn change (Sunday 2026-10-25): the instants move
	// by an hour, the wall clock does not.
	a := series(paris(2027, 3, 22, 9, 0), paris(2027, 3, 22, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{1},
	})
	windows := a.Occurrences(paris(2027, 3, 22, 0, 0), paris(2027, 3, 30, 0, 0))
	require.Len(t, windows, 2)
	assert.Equal(t, time.Date(2027, 3, 22, 8, 0, 0, 0, time.UTC), windows[0].Start.UTC(), "CET: UTC+1")
	assert.Equal(t, time.Date(2027, 3, 29, 7, 0, 0, 0, time.UTC), windows[1].Start.UTC(), "CEST: UTC+2")
	for _, w := range windows {
		local := w.Start.In(PracticeLocation)
		assert.Equal(t, 9, local.Hour())
		assert.Equal(t, 4*time.Hour, w.End.Sub(w.Start))
	}

	autumn := series(paris(2026, 10, 19, 9, 0), paris(2026, 10, 19, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{1},
	})
	windows = autumn.Occurrences(paris(2026, 10, 19, 0, 0), paris(2026, 10, 27, 0, 0))
	require.Len(t, windows, 2)
	assert.Equal(t, time.Date(2026, 10, 19, 7, 0, 0, 0, time.UTC), windows[0].Start.UTC(), "CEST: UTC+2")
	assert.Equal(t, time.Date(2026, 10, 26, 8, 0, 0, 0, time.UTC), windows[1].Start.UTC(), "CET: UTC+1")
}

func TestOccurrencesMonthlySkipsMonthsWithoutTheDay(t *testing.T) {
	// Monthly on the 31st from 2027-01-31: February, April and June have no
	// 31st and get no occurrence; the series never drifts to the 30th.
	a := series(paris(2027, 1, 31, 9, 0), paris(2027, 1, 31, 12, 0), RecurrencePattern{
		Type: RecurrenceMonthly, Interval: 1,
	})

	windows := a.Occurrences(paris(2027, 1, 1, 0, 0), paris(2027, 8, 1, 0, 0))

	assertInstants(t, []time.Time{
		paris(2027, 1, 31, 9, 0), paris(2027, 3, 31, 9, 0), paris(2027, 5, 31, 9, 0), paris(2027, 7, 31, 9, 0),
	}, occurrenceStarts(windows))
}

func TestOccurrencesMonthlyInterval(t *testing.T) {
	// Every second month on the 15th from 2026-11-15.
	a := series(paris(2026, 11, 15, 9, 0), paris(2026, 11, 15, 12, 0), RecurrencePattern{
		Type: RecurrenceMonthly, Interval: 2,
	})

	windows := a.Occurrences(paris(2026, 11, 1, 0, 0), paris(2027, 6, 1, 0, 0))

	assertInstants(t, []time.Time{
		paris(2026, 11, 15, 9, 0), paris(2027, 1, 15, 9, 0), paris(2027, 3, 15, 9, 0), paris(2027, 5, 15, 9, 0),
	}, occurrenceStarts(windows))
}

func TestOccurrencesIncludeOneRunningAtTheRangeStart(t *testing.T) {
	// A daily 09:00–13:00: a range from 11:00 still contains that day's
	// occurrence (the engine skips its slots before the range itself).
	a := series(paris(2026, 11, 2, 9, 0), paris(2026, 11, 2, 13, 0), RecurrencePattern{
		Type: RecurrenceDaily, Interval: 1,
	})

	windows := a.Occurrences(paris(2026, 11, 4, 11, 0), paris(2026, 11, 4, 12, 0))

	assertInstants(t, []time.Time{paris(2026, 11, 4, 9, 0)}, occurrenceStarts(windows))
}

func TestOccurrencesOneOff(t *testing.T) {
	a := &Availability{ID: uuid.New(), StartTime: paris(2026, 11, 2, 9, 0), EndTime: paris(2026, 11, 2, 13, 0)}

	assert.Len(t, a.Occurrences(paris(2026, 11, 2, 0, 0), paris(2026, 11, 3, 0, 0)), 1)
	assert.Empty(t, a.Occurrences(paris(2026, 11, 3, 0, 0), paris(2026, 11, 4, 0, 0)))
}
