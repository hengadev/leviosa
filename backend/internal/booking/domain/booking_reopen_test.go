package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBooking_MarkNoShow_SetsNoShowAt(t *testing.T) {
	b := bookingIn(t, BookingStatusConfirmed)
	require.NoError(t, b.MarkNoShow())
	require.NotNil(t, b.NoShowAt)
	assert.WithinDuration(t, time.Now(), *b.NoShowAt, time.Minute)
}

func TestBooking_Reopen(t *testing.T) {
	t.Run("puts a completed booking back to confirmed the same day", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		require.NoError(t, b.Complete())
		require.NoError(t, b.Reopen(time.Now()))
		assert.Equal(t, BookingStatusConfirmed, b.Status)
		assert.Nil(t, b.CompletedAt)
	})

	t.Run("puts a no-show booking back to confirmed the same day", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		require.NoError(t, b.MarkNoShow())
		require.NoError(t, b.Reopen(time.Now()))
		assert.Equal(t, BookingStatusConfirmed, b.Status)
		assert.Nil(t, b.NoShowAt)
	})

	t.Run("refuses the next day", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		require.NoError(t, b.Complete())
		nextDay := time.Now().AddDate(0, 0, 1)
		assert.ErrorIs(t, b.Reopen(nextDay), ErrReopenSameDayOnly)
		assert.Equal(t, BookingStatusCompleted, b.Status)
		assert.NotNil(t, b.CompletedAt)

		b = bookingIn(t, BookingStatusConfirmed)
		require.NoError(t, b.MarkNoShow())
		assert.ErrorIs(t, b.Reopen(nextDay), ErrReopenSameDayOnly)
		assert.Equal(t, BookingStatusNoShow, b.Status)
		assert.NotNil(t, b.NoShowAt)
	})

	for _, status := range []BookingStatus{BookingStatusAwaitingPayment, BookingStatusExpired, BookingStatusCancelled, BookingStatusConfirmed} {
		t.Run("refuses reopening a "+string(status)+" booking", func(t *testing.T) {
			b := bookingIn(t, status)
			assert.ErrorIs(t, b.Reopen(time.Now()), ErrBookingNotReopenable)
			assert.Equal(t, status, b.Status)
		})
	}

	t.Run("refuses a closed booking without its timestamp", func(t *testing.T) {
		b := bookingIn(t, BookingStatusCompleted)
		assert.ErrorIs(t, b.Reopen(time.Now()), ErrReopenSameDayOnly)
		assert.Equal(t, BookingStatusCompleted, b.Status)
	})
}
