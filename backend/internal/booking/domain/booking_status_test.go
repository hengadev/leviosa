package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Booking lifecycle (ADR-0018):
//
//	awaiting_payment → confirmed | expired
//	confirmed → completed | cancelled | no_show

func newAwaitingBooking(t *testing.T) *Booking {
	t.Helper()
	b, err := NewBooking(uuid.New(), nil, uuid.New(), uuid.New(), 5000, "EUR", "Jane", "Doe", "jane@example.com", "")
	require.NoError(t, err)
	return b
}

func bookingIn(t *testing.T, status BookingStatus) *Booking {
	t.Helper()
	b := newAwaitingBooking(t)
	b.Status = status
	return b
}

func TestNewBooking_StartsAwaitingPayment(t *testing.T) {
	b := newAwaitingBooking(t)
	assert.Equal(t, BookingStatusAwaitingPayment, b.Status)
	assert.Equal(t, PaymentStatusPending, b.PaymentStatus)
	assert.False(t, b.IsAppointment())
}

func TestBooking_Confirm(t *testing.T) {
	t.Run("moves a booking awaiting payment to confirmed", func(t *testing.T) {
		b := newAwaitingBooking(t)
		require.NoError(t, b.Confirm())
		assert.Equal(t, BookingStatusConfirmed, b.Status)
		assert.True(t, b.IsAppointment())
	})

	for _, status := range []BookingStatus{BookingStatusConfirmed, BookingStatusExpired, BookingStatusCancelled, BookingStatusCompleted, BookingStatusNoShow} {
		t.Run("rejects confirming a "+string(status)+" booking", func(t *testing.T) {
			b := bookingIn(t, status)
			assert.ErrorIs(t, b.Confirm(), ErrBookingNotAwaitingPayment)
			assert.Equal(t, status, b.Status)
		})
	}
}

func TestBooking_Expire(t *testing.T) {
	t.Run("moves a booking awaiting payment to expired", func(t *testing.T) {
		b := newAwaitingBooking(t)
		require.NoError(t, b.Expire())
		assert.Equal(t, BookingStatusExpired, b.Status)
		assert.False(t, b.IsAppointment())
	})

	for _, status := range []BookingStatus{BookingStatusConfirmed, BookingStatusExpired, BookingStatusCancelled, BookingStatusCompleted, BookingStatusNoShow} {
		t.Run("rejects expiring a "+string(status)+" booking", func(t *testing.T) {
			b := bookingIn(t, status)
			assert.ErrorIs(t, b.Expire(), ErrBookingNotAwaitingPayment)
			assert.Equal(t, status, b.Status)
		})
	}
}

func TestBooking_HoldUntil(t *testing.T) {
	t.Run("sets the hold of a booking awaiting payment", func(t *testing.T) {
		b := newAwaitingBooking(t)
		expiresAt := time.Now().Add(15 * time.Minute)
		require.NoError(t, b.HoldUntil(expiresAt))
		require.NotNil(t, b.HoldExpiresAt)
		assert.True(t, expiresAt.Equal(*b.HoldExpiresAt))
	})

	t.Run("rejects a hold on a confirmed booking", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		assert.ErrorIs(t, b.HoldUntil(time.Now()), ErrBookingNotAwaitingPayment)
		assert.Nil(t, b.HoldExpiresAt)
	})
}

func TestBooking_ConfirmedTransitionsRejectedBeforeConfirmation(t *testing.T) {
	for _, status := range []BookingStatus{BookingStatusAwaitingPayment, BookingStatusExpired} {
		t.Run("rejects cancelling a "+string(status)+" booking", func(t *testing.T) {
			b := bookingIn(t, status)
			assert.ErrorIs(t, b.Cancel("reason"), ErrBookingNotConfirmed)
			assert.Equal(t, status, b.Status)
			assert.Nil(t, b.CancelledAt)
		})
		t.Run("rejects completing a "+string(status)+" booking", func(t *testing.T) {
			b := bookingIn(t, status)
			assert.ErrorIs(t, b.Complete(), ErrBookingNotConfirmed)
			assert.Equal(t, status, b.Status)
		})
		t.Run("rejects marking a "+string(status)+" booking as no-show", func(t *testing.T) {
			b := bookingIn(t, status)
			assert.ErrorIs(t, b.MarkNoShow(), ErrBookingNotConfirmed)
			assert.Equal(t, status, b.Status)
		})
		t.Run("a "+string(status)+" booking is not cancellable", func(t *testing.T) {
			assert.False(t, bookingIn(t, status).IsCancellable())
		})
	}
}

func TestBooking_ConfirmedTransitionsUnchanged(t *testing.T) {
	t.Run("a confirmed booking completes", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		require.NoError(t, b.Complete())
		assert.Equal(t, BookingStatusCompleted, b.Status)
	})
	t.Run("a confirmed booking cancels", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		require.NoError(t, b.Cancel("reason"))
		assert.Equal(t, BookingStatusCancelled, b.Status)
	})
	t.Run("a confirmed booking is marked no-show", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		require.NoError(t, b.MarkNoShow())
		assert.Equal(t, BookingStatusNoShow, b.Status)
	})
}

func TestBooking_OccupiesTime(t *testing.T) {
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)

	t.Run("a booking awaiting payment occupies time while its hold runs", func(t *testing.T) {
		b := newAwaitingBooking(t)
		require.NoError(t, b.HoldUntil(now.Add(time.Minute)))
		assert.True(t, b.OccupiesTime(now))
	})

	t.Run("a booking awaiting payment frees its slot once the hold elapsed", func(t *testing.T) {
		b := newAwaitingBooking(t)
		require.NoError(t, b.HoldUntil(now))
		assert.False(t, b.OccupiesTime(now), "the hold ends at its expiry")
		assert.False(t, b.OccupiesTime(now.Add(time.Second)))
	})

	t.Run("a booking awaiting payment without a hold keeps its slot", func(t *testing.T) {
		assert.True(t, newAwaitingBooking(t).OccupiesTime(now))
	})

	for status, occupies := range map[BookingStatus]bool{
		BookingStatusConfirmed: true,
		BookingStatusCompleted: true,
		BookingStatusNoShow:    true,
		BookingStatusCancelled: false,
		BookingStatusExpired:   false,
	} {
		t.Run(string(status), func(t *testing.T) {
			assert.Equal(t, occupies, bookingIn(t, status).OccupiesTime(now))
		})
	}
}

func TestParseBookingLocale(t *testing.T) {
	for input, want := range map[string]BookingLocale{"": BookingLocaleFR, "fr": BookingLocaleFR, "en": BookingLocaleEN} {
		got, err := ParseBookingLocale(input)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
	_, err := ParseBookingLocale("de")
	assert.ErrorIs(t, err, ErrInvalidBookingLocale)
}

func TestBooking_ConfirmLatePayment(t *testing.T) {
	for _, status := range []BookingStatus{BookingStatusAwaitingPayment, BookingStatusExpired} {
		t.Run("confirms a "+string(status)+" booking", func(t *testing.T) {
			b := bookingIn(t, status)
			require.NoError(t, b.ConfirmLatePayment())
			assert.Equal(t, BookingStatusConfirmed, b.Status)
			assert.True(t, b.IsAppointment())
		})
	}
	for _, status := range []BookingStatus{BookingStatusConfirmed, BookingStatusCancelled, BookingStatusCompleted, BookingStatusNoShow} {
		t.Run("rejects confirming a "+string(status)+" booking", func(t *testing.T) {
			b := bookingIn(t, status)
			assert.ErrorIs(t, b.ConfirmLatePayment(), ErrBookingNotAwaitingPayment)
			assert.Equal(t, status, b.Status)
		})
	}
}

func TestBooking_HoldElapsed(t *testing.T) {
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	held := func(status BookingStatus, expiresAt *time.Time) *Booking {
		b := bookingIn(t, status)
		b.HoldExpiresAt = expiresAt
		return b
	}
	past, future := now.Add(-time.Minute), now.Add(time.Minute)

	assert.True(t, held(BookingStatusAwaitingPayment, &past).HoldElapsed(now))
	assert.True(t, held(BookingStatusAwaitingPayment, &now).HoldElapsed(now), "the hold ends at its expiry")
	assert.False(t, held(BookingStatusAwaitingPayment, &future).HoldElapsed(now))
	assert.False(t, held(BookingStatusAwaitingPayment, nil).HoldElapsed(now), "a hold without an expiry never elapses")
	assert.False(t, held(BookingStatusConfirmed, &past).HoldElapsed(now))
}

func TestBooking_OverlapsAny(t *testing.T) {
	start := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	b := newAwaitingBooking(t)
	b.SlotStartTime, b.SlotEndTime = start, start.Add(time.Hour)
	buffer := 15 * time.Minute

	assert.False(t, b.OverlapsAny(buffer, nil))
	assert.True(t, b.OverlapsAny(buffer, []SlotOccupancy{{Start: start.Add(30 * time.Minute), End: start.Add(2 * time.Hour)}}))
	assert.True(t, b.OverlapsAny(buffer, []SlotOccupancy{{Start: start.Add(70 * time.Minute), End: start.Add(2 * time.Hour)}}),
		"the booking's own buffer counts")
	assert.False(t, b.OverlapsAny(buffer, []SlotOccupancy{{Start: start.Add(75 * time.Minute), End: start.Add(2 * time.Hour)}}),
		"touching intervals do not overlap")
	assert.False(t, b.OverlapsAny(buffer, []SlotOccupancy{{Start: start.Add(-time.Hour), End: start}}))
}
