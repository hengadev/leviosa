package booking

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
)

// planFinancialRows is the pure aggregation behind GetFinancialSummary:
// online money on the booking's creation day, on-site money on its
// recording day (Europe/Paris). These tests cover the on-site side and the
// month boundaries; the online side is covered by the integration tests.
func TestPlanFinancialRows(t *testing.T) {
	// November 2026, UTC midnights as the HTTP layer produces them.
	novFrom := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	novTo := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	octFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	octTo := novFrom

	paid := func(status domain.PaymentStatus) *domain.Booking {
		return &domain.Booking{
			ID:              uuid.New(),
			Status:          domain.BookingStatusCompleted,
			PaymentStatus:   status,
			TotalPriceCents: 12000,
		}
	}

	t.Run("on-site payment counts on its recording day, not the creation day", func(t *testing.T) {
		b := paid(domain.PaymentStatusPaid)
		b.CreatedAt = time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC) // October
		deposit := 2000
		b.AmountPaidOnlineCents = &deposit
		recorded := time.Date(2026, 11, 20, 15, 0, 0, 0, time.UTC)
		b.AmountPaidOnSiteCents = &[]int{10000}[0]
		b.PaidOnSiteAt = &recorded

		// November: only the on-site row.
		summary, rows := planFinancialRows([]*domain.Booking{b}, novFrom, novTo)
		assert.Equal(t, 10000, summary.OnSiteCents)
		assert.Equal(t, 10000, summary.GrossRevenueCents)
		assert.Equal(t, 10000, summary.NetRevenueCents)
		require.Len(t, rows, 1)
		assert.Equal(t, domain.PaymentModeOnSite, rows[0].mode)
		assert.Equal(t, 10000, rows[0].amountCents)
		assert.Equal(t, recorded, rows[0].paidAt)

		// October: only the online deposit row.
		octSummary, octRows := planFinancialRows([]*domain.Booking{b}, octFrom, octTo)
		assert.Equal(t, 0, octSummary.OnSiteCents)
		assert.Equal(t, deposit, octSummary.GrossRevenueCents)
		require.Len(t, octRows, 1)
		assert.Equal(t, domain.PaymentModeOnline, octRows[0].mode)
		assert.Equal(t, deposit, octRows[0].amountCents)
	})

	t.Run("gross and net include on-site alongside the online deposit", func(t *testing.T) {
		b := paid(domain.PaymentStatusPaid)
		b.CreatedAt = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
		deposit := 2000
		b.AmountPaidOnlineCents = &deposit
		recorded := time.Date(2026, 11, 10, 11, 0, 0, 0, time.UTC)
		b.AmountPaidOnSiteCents = &[]int{10000}[0]
		b.PaidOnSiteAt = &recorded

		summary, rows := planFinancialRows([]*domain.Booking{b}, novFrom, novTo)
		assert.Equal(t, 12000, summary.GrossRevenueCents)
		assert.Equal(t, 10000, summary.OnSiteCents)
		assert.Equal(t, 12000, summary.NetRevenueCents)
		require.Len(t, rows, 2)
		// Most recent payment first: the on-site payment came after the deposit.
		assert.Equal(t, domain.PaymentModeOnSite, rows[0].mode)
		assert.Equal(t, domain.PaymentModeOnline, rows[1].mode)
	})

	t.Run("an undone payment disappears (no row, no total)", func(t *testing.T) {
		b := paid(domain.PaymentStatusPaid)
		b.CreatedAt = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
		deposit := 2000
		b.AmountPaidOnlineCents = &deposit
		// UndoOnSitePayment clears both fields; nothing is left to count.
		b.AmountPaidOnSiteCents = nil
		b.PaidOnSiteAt = nil

		summary, rows := planFinancialRows([]*domain.Booking{b}, novFrom, novTo)
		assert.Equal(t, 0, summary.OnSiteCents)
		assert.Equal(t, deposit, summary.GrossRevenueCents)
		require.Len(t, rows, 1)
		assert.Equal(t, domain.PaymentModeOnline, rows[0].mode)
	})

	t.Run("a recording just before the Paris midnight of the first day is out", func(t *testing.T) {
		b := paid(domain.PaymentStatusPaid)
		b.CreatedAt = time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
		b.AmountPaidOnSiteCents = &[]int{10000}[0]
		// 2026-11-01 00:30 Paris is still 2026-10-31 in UTC: October's day.
		octoberNight := time.Date(2026, 10, 31, 22, 30, 0, 0, time.UTC)
		b.PaidOnSiteAt = &octoberNight

		novSummary, novRows := planFinancialRows([]*domain.Booking{b}, novFrom, novTo)
		assert.Equal(t, 0, novSummary.OnSiteCents)
		assert.Empty(t, novRows)

		octSummary, octRows := planFinancialRows([]*domain.Booking{b}, octFrom, octTo)
		assert.Equal(t, 10000, octSummary.OnSiteCents)
		require.Len(t, octRows, 1)
		assert.Equal(t, domain.PaymentModeOnSite, octRows[0].mode)
	})

	t.Run("a disputed online payment stays out of gross and net", func(t *testing.T) {
		b := paid(domain.PaymentStatusDisputed)
		b.CreatedAt = time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
		deposit := 2000
		b.AmountPaidOnlineCents = &deposit

		summary, rows := planFinancialRows([]*domain.Booking{b}, novFrom, novTo)
		assert.Equal(t, deposit, summary.DisputedCents)
		assert.Equal(t, 0, summary.GrossRevenueCents)
		assert.Equal(t, 0, summary.NetRevenueCents)
		require.Len(t, rows, 1)
		assert.Equal(t, domain.PaymentModeOnline, rows[0].mode)
	})
}
