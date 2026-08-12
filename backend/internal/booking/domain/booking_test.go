package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPaidBooking(t *testing.T) *Booking {
	t.Helper()
	b, err := NewBooking(uuid.New(), nil, uuid.New(), uuid.New(), 5000, "EUR", "Jane", "Doe", "jane@example.com", "")
	require.NoError(t, err)
	require.NoError(t, b.MarkPaymentPaid())
	return b
}

func TestBooking_MarkPaymentDisputed(t *testing.T) {
	t.Run("transitions a paid booking to disputed", func(t *testing.T) {
		b := newTestPaidBooking(t)

		err := b.MarkPaymentDisputed()

		require.NoError(t, err)
		assert.Equal(t, PaymentStatusDisputed, b.PaymentStatus)
	})

	t.Run("rejects disputing a pending (unpaid) booking", func(t *testing.T) {
		b, err := NewBooking(uuid.New(), nil, uuid.New(), uuid.New(), 5000, "EUR", "Jane", "Doe", "jane@example.com", "")
		require.NoError(t, err)
		require.Equal(t, PaymentStatusPending, b.PaymentStatus)

		err = b.MarkPaymentDisputed()

		assert.ErrorIs(t, err, ErrCannotDisputeUnpaidBooking)
		assert.Equal(t, PaymentStatusPending, b.PaymentStatus, "status must not change on error")
	})

	t.Run("rejects disputing an already-refunded booking", func(t *testing.T) {
		b := newTestPaidBooking(t)
		require.NoError(t, b.RefundPayment())

		err := b.MarkPaymentDisputed()

		assert.ErrorIs(t, err, ErrCannotDisputeUnpaidBooking)
		assert.Equal(t, PaymentStatusRefunded, b.PaymentStatus)
	})

	t.Run("rejects disputing a booking that is already disputed", func(t *testing.T) {
		b := newTestPaidBooking(t)
		require.NoError(t, b.MarkPaymentDisputed())

		err := b.MarkPaymentDisputed()

		assert.ErrorIs(t, err, ErrCannotDisputeUnpaidBooking)
	})
}

func TestBooking_ResolveDispute(t *testing.T) {
	t.Run("reverts to paid when the dispute is won", func(t *testing.T) {
		b := newTestPaidBooking(t)
		require.NoError(t, b.MarkPaymentDisputed())

		err := b.ResolveDispute("won")

		require.NoError(t, err)
		assert.Equal(t, PaymentStatusPaid, b.PaymentStatus)
	})

	t.Run("reverts to paid for any non-lost terminal status", func(t *testing.T) {
		b := newTestPaidBooking(t)
		require.NoError(t, b.MarkPaymentDisputed())

		err := b.ResolveDispute("warning_closed")

		require.NoError(t, err)
		assert.Equal(t, PaymentStatusPaid, b.PaymentStatus)
	})

	t.Run("reverts to refunded when the dispute is lost", func(t *testing.T) {
		b := newTestPaidBooking(t)
		require.NoError(t, b.MarkPaymentDisputed())

		err := b.ResolveDispute("lost")

		require.NoError(t, err)
		assert.Equal(t, PaymentStatusRefunded, b.PaymentStatus)
	})

	t.Run("rejects resolving a booking that is not disputed", func(t *testing.T) {
		b := newTestPaidBooking(t)

		err := b.ResolveDispute("won")

		assert.ErrorIs(t, err, ErrCannotResolveNonDisputedBooking)
		assert.Equal(t, PaymentStatusPaid, b.PaymentStatus)
	})
}
