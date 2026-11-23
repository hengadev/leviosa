package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startedBooking returns a booking of 50 € in the given status whose session
// started an hour ago, with paidOnline cents paid online (nil: no amount
// recorded).
func startedBooking(t *testing.T, status BookingStatus, paidOnline *int) *Booking {
	t.Helper()
	b := bookingIn(t, status)
	b.SlotStartTime = time.Now().Add(-time.Hour)
	b.SlotEndTime = b.SlotStartTime.Add(time.Hour)
	b.AmountPaidOnlineCents = paidOnline
	if paidOnline != nil && *paidOnline > 0 {
		b.PaymentStatus = PaymentStatusPaid
	}
	return b
}

func cents(v int) *int { return &v }

func TestBooking_AmountDueOnSiteCents(t *testing.T) {
	t.Run("is 0 for a full online payment", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, cents(5000))
		assert.Equal(t, 0, b.AmountDueOnSiteCents())
	})

	t.Run("is 0 for a booking paid online before the amount was recorded", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, nil)
		b.PaymentStatus = PaymentStatusPaid
		assert.Equal(t, 0, b.AmountDueOnSiteCents())
	})

	t.Run("is the price minus the deposit", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, cents(2000))
		assert.Equal(t, 3000, b.AmountDueOnSiteCents())
	})

	t.Run("is the full price when nothing was paid online", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, cents(0))
		assert.Equal(t, 5000, b.AmountDueOnSiteCents())
	})
}

func TestBooking_RecordOnSitePayment(t *testing.T) {
	for _, status := range []BookingStatus{BookingStatusConfirmed, BookingStatusCompleted} {
		t.Run("records the amount due on a started "+string(status)+" booking", func(t *testing.T) {
			b := startedBooking(t, status, cents(2000))
			now := time.Now()
			require.NoError(t, b.RecordOnSitePayment(now))
			require.NotNil(t, b.AmountPaidOnSiteCents)
			assert.Equal(t, 3000, *b.AmountPaidOnSiteCents)
			require.NotNil(t, b.PaidOnSiteAt)
			assert.True(t, b.PaidOnSiteAt.Equal(now))
			assert.Equal(t, status, b.Status)
		})
	}

	t.Run("records the full price when nothing was paid online", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, cents(0))
		require.NoError(t, b.RecordOnSitePayment(time.Now()))
		assert.Equal(t, 5000, *b.AmountPaidOnSiteCents)
	})

	for _, status := range []BookingStatus{BookingStatusCancelled, BookingStatusNoShow, BookingStatusAwaitingPayment, BookingStatusExpired} {
		t.Run("refuses a "+string(status)+" booking", func(t *testing.T) {
			b := startedBooking(t, status, cents(2000))
			assert.ErrorIs(t, b.RecordOnSitePayment(time.Now()), ErrOnSitePaymentNotAllowed)
			assert.Nil(t, b.AmountPaidOnSiteCents)
			assert.Nil(t, b.PaidOnSiteAt)
		})
	}

	t.Run("refuses before the session starts", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, cents(2000))
		b.SlotStartTime = time.Now().Add(time.Hour)
		assert.ErrorIs(t, b.RecordOnSitePayment(time.Now()), ErrOnSitePaymentBeforeStart)
		assert.Nil(t, b.AmountPaidOnSiteCents)
	})

	t.Run("refuses when nothing is due", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, cents(5000))
		assert.ErrorIs(t, b.RecordOnSitePayment(time.Now()), ErrNothingDueOnSite)
		assert.Nil(t, b.AmountPaidOnSiteCents)
	})

	t.Run("refuses when already recorded", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, cents(2000))
		first := time.Now()
		require.NoError(t, b.RecordOnSitePayment(first))
		assert.ErrorIs(t, b.RecordOnSitePayment(first.Add(time.Minute)), ErrOnSitePaymentAlreadyRecorded)
		assert.True(t, b.PaidOnSiteAt.Equal(first))
	})
}

func TestBooking_UndoOnSitePayment(t *testing.T) {
	t.Run("undoes the payment the same day", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, cents(2000))
		require.NoError(t, b.RecordOnSitePayment(time.Now()))
		require.NoError(t, b.UndoOnSitePayment(time.Now()))
		assert.Nil(t, b.AmountPaidOnSiteCents)
		assert.Nil(t, b.PaidOnSiteAt)
	})

	t.Run("refuses the next day", func(t *testing.T) {
		b := startedBooking(t, BookingStatusCompleted, cents(2000))
		require.NoError(t, b.RecordOnSitePayment(time.Now()))
		assert.ErrorIs(t, b.UndoOnSitePayment(time.Now().AddDate(0, 0, 1)), ErrOnSitePaymentUndoSameDayOnly)
		require.NotNil(t, b.AmountPaidOnSiteCents)
		assert.Equal(t, 3000, *b.AmountPaidOnSiteCents)
	})

	t.Run("uses the Paris day, not the UTC one", func(t *testing.T) {
		// 23:30 in Paris and 00:10 the next morning in Paris are two days,
		// whatever the UTC dates.
		recorded := time.Date(2026, 11, 23, 23, 30, 0, 0, PracticeLocation)
		b := startedBooking(t, BookingStatusConfirmed, cents(2000))
		b.SlotStartTime = recorded.Add(-time.Hour)
		require.NoError(t, b.RecordOnSitePayment(recorded))
		assert.ErrorIs(t, b.UndoOnSitePayment(recorded.Add(40*time.Minute)), ErrOnSitePaymentUndoSameDayOnly)
		assert.NoError(t, b.UndoOnSitePayment(recorded.Add(20*time.Minute)))
	})

	t.Run("refuses when nothing is recorded", func(t *testing.T) {
		b := startedBooking(t, BookingStatusConfirmed, cents(2000))
		assert.ErrorIs(t, b.UndoOnSitePayment(time.Now()), ErrOnSitePaymentNotRecorded)
	})
}
