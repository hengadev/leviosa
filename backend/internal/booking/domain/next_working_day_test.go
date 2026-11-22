package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Next working day (issue 037): the first Paris date after today with an
// open occurrence or an appointment booking.

func appointmentOn(status BookingStatus, start time.Time) *Booking {
	return &Booking{
		ID:            uuid.New(),
		Status:        status,
		SlotStartTime: start,
		SlotEndTime:   start.Add(time.Hour),
	}
}

func TestNextWorkingDayWeeklyAcrossWeekend(t *testing.T) {
	// Every Tuesday 09:00–13:00 from Tuesday 2026-11-03. Today is Sunday
	// 2026-11-08: the next open day is Tuesday the 10th.
	a := series(paris(2026, 11, 3, 9, 0), paris(2026, 11, 3, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{2},
	})

	day, ok := NextWorkingDay([]*Availability{a}, nil, nil, paris(2026, 11, 8, 20, 0), 60)

	require.True(t, ok)
	assert.Equal(t, paris(2026, 11, 10, 0, 0), day)
}

func TestNextWorkingDayExceptionClosesNextOpenDay(t *testing.T) {
	// Every Tuesday, but Tuesday 2026-11-10 is closed by an exception: from
	// Sunday 2026-11-08 the next open day is Tuesday the 17th.
	a := series(paris(2026, 11, 3, 9, 0), paris(2026, 11, 3, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{2},
	})
	exceptions := []AvailabilityException{exceptionOn(a, date(2026, 11, 10), date(2026, 11, 10))}

	day, ok := NextWorkingDay([]*Availability{a}, exceptions, nil, paris(2026, 11, 8, 20, 0), 60)

	require.True(t, ok)
	assert.Equal(t, paris(2026, 11, 17, 0, 0), day)
}

func TestNextWorkingDayOneOff(t *testing.T) {
	// A one-off availability on Saturday 2026-11-14.
	a := &Availability{ID: uuid.New(), StartTime: paris(2026, 11, 14, 9, 0), EndTime: paris(2026, 11, 14, 13, 0), Status: AvailabilityStatusAvailable}

	day, ok := NextWorkingDay([]*Availability{a}, nil, nil, paris(2026, 11, 13, 22, 0), 60)

	require.True(t, ok)
	assert.Equal(t, paris(2026, 11, 14, 0, 0), day)
}

func TestNextWorkingDayBookingWithoutAvailability(t *testing.T) {
	// A confirmed booking on Wednesday 2026-11-11, no availability at all.
	booking := appointmentOn(BookingStatusConfirmed, paris(2026, 11, 11, 10, 0))

	day, ok := NextWorkingDay(nil, nil, []*Booking{booking}, paris(2026, 11, 10, 23, 0), 60)

	require.True(t, ok)
	assert.Equal(t, paris(2026, 11, 11, 0, 0), day)
}

func TestNextWorkingDayNothingWithinLookAhead(t *testing.T) {
	// Every Tuesday; a look-ahead of 2 days from Saturday 2026-11-07 only
	// reaches Monday the 9th, before the next Tuesday.
	a := series(paris(2026, 11, 3, 9, 0), paris(2026, 11, 3, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{2},
	})

	_, ok := NextWorkingDay([]*Availability{a}, nil, nil, paris(2026, 11, 7, 10, 0), 2)

	assert.False(t, ok)
}

func TestNextWorkingDayNeverToday(t *testing.T) {
	// A daily availability and a booking today: today itself never counts,
	// not even at 23:59.
	a := series(paris(2026, 11, 2, 9, 0), paris(2026, 11, 2, 13, 0), RecurrencePattern{Type: RecurrenceDaily, Interval: 1})
	today := paris(2026, 11, 10, 23, 59)
	booking := appointmentOn(BookingStatusConfirmed, paris(2026, 11, 10, 23, 30))

	day, ok := NextWorkingDay([]*Availability{a}, nil, []*Booking{booking}, today, 60)

	require.True(t, ok)
	assert.Equal(t, paris(2026, 11, 11, 0, 0), day)
}

func TestNextWorkingDayExcludesNonWorkingBookings(t *testing.T) {
	// Cancelled, awaiting-payment and expired bookings do not make a day
	// worked; the no-show one on Thursday does.
	bookings := []*Booking{
		appointmentOn(BookingStatusCancelled, paris(2026, 11, 11, 10, 0)),
		appointmentOn(BookingStatusAwaitingPayment, paris(2026, 11, 11, 11, 0)),
		appointmentOn(BookingStatusExpired, paris(2026, 11, 11, 12, 0)),
		appointmentOn(BookingStatusNoShow, paris(2026, 11, 12, 9, 0)),
	}

	day, ok := NextWorkingDay(nil, nil, bookings, paris(2026, 11, 10, 20, 0), 60)

	require.True(t, ok)
	assert.Equal(t, paris(2026, 11, 12, 0, 0), day)
}

func TestNextWorkingDayIgnoresUnavailableAvailabilities(t *testing.T) {
	// A cancelled availability does not make its days worked.
	a := series(paris(2026, 11, 3, 9, 0), paris(2026, 11, 3, 13, 0), RecurrencePattern{
		Type: RecurrenceWeekly, Interval: 1, DaysOfWeek: []int{2},
	})
	a.Status = AvailabilityStatusCancelled

	_, ok := NextWorkingDay([]*Availability{a}, nil, nil, paris(2026, 11, 7, 10, 0), 10)

	assert.False(t, ok)
}
