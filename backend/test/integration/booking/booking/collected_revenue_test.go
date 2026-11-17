package booking

import (
	"context"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Finance figures count the money collected, on the day it was paid: a
// deposit paid today for a session to come counts today, for the deposit,
// not the full price.
//
// make test-func TEST_NAME=TestCollectedRevenue TEST_PATH=test/integration/booking/booking/collected_revenue_test.go
func TestCollectedRevenue(t *testing.T) {
	ctx := context.Background()
	payments := mockPayments(t)

	t.Run("should count a deposit paid today for a session to come", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		held := holdGuestBooking(t, ctx, f, f.at(10, 15), "claire@example.com", domain.BookingLocaleFR)
		require.True(t, held.SlotStartTime.After(time.Now()), "the session is still to come")
		const depositCents = 2000
		require.Less(t, depositCents, held.TotalPriceCents, "a deposit, not the full price")

		require.NoError(t, payments.ConfirmPaymentIntent(ctx, *held.PaymentIntentID))
		event := paymentSucceededEvent(t, ctx, held.ID)
		event.Amount = depositCents
		require.NoError(t, service.HandlePaymentWebhook(ctx, event))
		require.Equal(t, domain.BookingStatusConfirmed, storedBooking(t, ctx, held.ID).Status)

		// The practitioner's earnings.
		earnings, err := service.GetPartnerEarnings(ctx, held.PartnerID)
		require.NoError(t, err)
		assert.Equal(t, depositCents, earnings.CurrentMonthCents)
		assert.Equal(t, depositCents, earnings.NextPayoutCents)
		require.Len(t, earnings.Transactions, 1)
		tx := earnings.Transactions[0]
		assert.Equal(t, depositCents, tx.AmountCents)
		assert.Equal(t, depositCents, tx.NetCents)
		paidAt, err := time.Parse(time.RFC3339, tx.PaidAt)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), paidAt, time.Minute)

		// The admin's accounting for this month so far.
		now := time.Now().UTC()
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		summary, err := service.GetFinancialSummary(ctx, from, to)
		require.NoError(t, err)
		assert.Equal(t, depositCents, summary.Summary.GrossRevenueCents)
		assert.Equal(t, depositCents, summary.Summary.NetRevenueCents)
		require.Len(t, summary.Transactions, 1)
		assert.Equal(t, depositCents, summary.Transactions[0].AmountCents)

		// The admin's analytics for this month.
		analytics, err := service.GetAnalyticsSummary(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, depositCents, analytics.CurrentMonth.RevenueCents)
	})

	t.Run("should count nothing for a booking still awaiting payment", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotFixture(t, ctx)

		held := holdGuestBooking(t, ctx, f, f.at(10, 15), "claire@example.com", domain.BookingLocaleFR)

		earnings, err := service.GetPartnerEarnings(ctx, held.PartnerID)
		require.NoError(t, err)
		assert.Equal(t, 0, earnings.CurrentMonthCents)
		assert.Empty(t, earnings.Transactions)
	})
}
