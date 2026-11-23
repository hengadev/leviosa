package booking

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestOnSiteFinancialSummary TEST_PATH=test/integration/booking/booking/on_site_financial_summary_test.go

// Financial figures count an on-site payment on the day it was recorded,
// apart from the online deposit that counts on the booking's creation day
// (issue 040).
func TestOnSiteFinancialSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("should count a recorded on-site payment next to the online deposit", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)
		booking := startSessionWithDeposit(t, ctx, bookingID)
		due := booking.TotalPriceCents - onSiteDepositCents

		resp := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		from, to := monthSoFar()
		summary, err := service.GetFinancialSummary(ctx, from, to)
		require.NoError(t, err)
		assert.Equal(t, onSiteDepositCents+due, summary.Summary.GrossRevenueCents)
		assert.Equal(t, due, summary.Summary.OnSiteCents)
		assert.Equal(t, onSiteDepositCents+due, summary.Summary.NetRevenueCents)
		require.Len(t, summary.Transactions, 2)

		var online, onSite *domain.FinancialTransaction
		for i := range summary.Transactions {
			switch summary.Transactions[i].PaymentMode {
			case domain.PaymentModeOnline:
				online = &summary.Transactions[i]
			case domain.PaymentModeOnSite:
				onSite = &summary.Transactions[i]
			}
		}
		require.NotNil(t, online, "the online deposit row")
		assert.Equal(t, onSiteDepositCents, online.AmountCents)
		require.NotNil(t, onSite, "the on-site row")
		assert.Equal(t, due, onSite.AmountCents)
		stored := storedOnSiteBooking(t, ctx, bookingID)
		require.NotNil(t, stored.PaidOnSiteAt)
		assert.WithinDuration(t, *stored.PaidOnSiteAt, onSite.PaidAt, time.Minute)

		// The CSV export carries the mode and the on-site total.
		adminToken := tu.SetupAdminUser(t, ctx, authCtx)
		req, err := http.NewRequest(http.MethodGet,
			testServerURL+"/admin/bookings/financial-summary?from="+from.Format(time.DateOnly)+"&to="+to.AddDate(0, 0, -1).Format(time.DateOnly)+"&format=csv", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: adminToken})
		csvResp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
		require.NoError(t, err)
		defer csvResp.Body.Close()
		require.Equal(t, http.StatusOK, csvResp.StatusCode)
		csvBody, err := io.ReadAll(csvResp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(csvBody), "on_site_cents,"+strconv.Itoa(due))
		assert.Contains(t, string(csvBody), "Sur place")
		assert.Contains(t, string(csvBody), "En ligne")
	})

	t.Run("should split a deposit from a later month and its on-site balance", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)
		booking := startSessionWithDeposit(t, ctx, bookingID)
		due := booking.TotalPriceCents - onSiteDepositCents

		resp := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// The deposit was paid online last month: move the booking's
		// creation back (a plain column) while the on-site payment stays
		// recorded today.
		today := practiceToday()
		firstOfThisMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastMonth := firstOfThisMonth.AddDate(0, 0, -15)
		_, err := testPool.Exec(ctx,
			"UPDATE booking.bookings SET created_at = $2 WHERE id = $1",
			bookingID, lastMonth)
		require.NoError(t, err)

		// Last month: the online deposit only.
		prevFrom := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		prevTo := firstOfThisMonth
		prev, err := service.GetFinancialSummary(ctx, prevFrom, prevTo)
		require.NoError(t, err)
		assert.Equal(t, onSiteDepositCents, prev.Summary.GrossRevenueCents)
		assert.Equal(t, 0, prev.Summary.OnSiteCents)
		require.Len(t, prev.Transactions, 1)
		assert.Equal(t, domain.PaymentModeOnline, prev.Transactions[0].PaymentMode)
		assert.Equal(t, onSiteDepositCents, prev.Transactions[0].AmountCents)

		// This month: the on-site balance only.
		from, to := monthSoFar()
		cur, err := service.GetFinancialSummary(ctx, from, to)
		require.NoError(t, err)
		assert.Equal(t, due, cur.Summary.GrossRevenueCents)
		assert.Equal(t, due, cur.Summary.OnSiteCents)
		require.Len(t, cur.Transactions, 1)
		assert.Equal(t, domain.PaymentModeOnSite, cur.Transactions[0].PaymentMode)
		assert.Equal(t, due, cur.Transactions[0].AmountCents)
	})

	t.Run("should drop the figures of an undone on-site payment", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)

		bookingID, partnerToken := setupClosedBooking(t)
		startSessionWithDeposit(t, ctx, bookingID)
		resp := onSitePaymentRequest(t, http.MethodPost, bookingID, partnerToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		undo := onSitePaymentRequest(t, http.MethodDelete, bookingID, partnerToken)
		defer undo.Body.Close()
		require.Equal(t, http.StatusOK, undo.StatusCode)

		from, to := monthSoFar()
		summary, err := service.GetFinancialSummary(ctx, from, to)
		require.NoError(t, err)
		assert.Equal(t, onSiteDepositCents, summary.Summary.GrossRevenueCents)
		assert.Equal(t, 0, summary.Summary.OnSiteCents)
		require.Len(t, summary.Transactions, 1)
		assert.Equal(t, domain.PaymentModeOnline, summary.Transactions[0].PaymentMode)
	})
}
