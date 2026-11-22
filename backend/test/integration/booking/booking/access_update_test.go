package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"
	tbooking "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/booking"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// accessUpdateCall calls GET or POST /admin/bookings/access-update with the
// given session token ("" for anonymous) and decodes the counts on 200.
func accessUpdateCall(t *testing.T, method, token string) (int, domain.AccessUpdateRecipients) {
	t.Helper()
	req, err := http.NewRequest(method, testServerURL+"/admin/bookings/access-update", nil)
	require.NoError(t, err)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: token})
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var counts domain.AccessUpdateRecipients
	if resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&counts))
	}
	return resp.StatusCode, counts
}

// waitForAccessUpdates waits until n access update emails are recorded, then
// gives an extra one the chance to show up before returning them all.
func waitForAccessUpdates(t *testing.T, n int) []notificationDomain.BookingAccessUpdateRequest {
	t.Helper()
	if n > 0 {
		require.Eventually(t, func() bool {
			return len(notificationEmailSpy.AccessUpdateSnapshot()) >= n
		}, 5*time.Second, 20*time.Millisecond, "expected %d access update email(s)", n)
	}
	time.Sleep(300 * time.Millisecond)
	return notificationEmailSpy.AccessUpdateSnapshot()
}

// The access update emails the clients of upcoming confirmed bookings only,
// each in its locale with the current door code, through the real
// notification adapter; the counts match; no door code is refused; the
// audit entry carries the counts, never the code (issue 026).
func TestAccessUpdate_EmailsUpcomingConfirmedBookingsOnly(t *testing.T) {
	ctx := context.Background()
	defer tu.ClearAuthData(t, ctx, authCtx)
	f := setupReminderFixture(t, ctx, "partner-access-update@example.com")
	t.Cleanup(func() { setPracticeAccess(t, "", "") })

	french := f.book(t, ctx, 30*time.Hour, "fr.access@example.com", domain.BookingLocaleFR)
	confirmBookingPayment(t, ctx, french.ID)
	english := f.book(t, ctx, 20*time.Hour, "en.access@example.com", domain.BookingLocaleEN)
	confirmBookingPayment(t, ctx, english.ID)

	// Started by the time of the send (the clock is moved 2 hours ahead).
	started := f.book(t, ctx, time.Hour, "started.access@example.com", domain.BookingLocaleFR)
	confirmBookingPayment(t, ctx, started.ID)

	for i, status := range []domain.BookingStatus{
		domain.BookingStatusCancelled, domain.BookingStatusCompleted, domain.BookingStatusNoShow,
	} {
		b := f.book(t, ctx, time.Duration(33+3*i)*time.Hour, string(status)+".access@example.com", domain.BookingLocaleFR)
		confirmBookingPayment(t, ctx, b.ID)
		require.NoError(t, tbooking.UpdateBookingStatus(t, ctx, testPool, b.ID, status))
	}
	expired := f.book(t, ctx, 42*time.Hour, "expired.access@example.com", domain.BookingLocaleFR)
	require.NoError(t, tbooking.UpdateBookingStatus(t, ctx, testPool, expired.ID, domain.BookingStatusExpired))
	f.book(t, ctx, 45*time.Hour, "awaiting.access@example.com", domain.BookingLocaleFR) // awaiting payment

	suiteClockOffset.Store(int64(2 * time.Hour))
	defer suiteClockOffset.Store(0)

	adminToken := tu.SetupAdminUser(t, ctx, authCtx)

	status, counts := accessUpdateCall(t, http.MethodGet, adminToken)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, domain.AccessUpdateRecipients{Recipients: 2, WithoutEmail: 0}, counts)

	// No door code set: refused, nothing sent.
	setPracticeAccess(t, "", "")
	notificationEmailSpy.Reset()
	status, _ = accessUpdateCall(t, http.MethodPost, adminToken)
	assert.Equal(t, http.StatusBadRequest, status)

	// The real adapter, reading the practice access at send time, sends the
	// emails the spy hands it.
	stop := notificationSpy.ForwardConfirmationsTo(accessNotificationAdapter)
	defer stop()
	assert.Empty(t, waitForAccessUpdates(t, 0), "no email without a door code")

	setPracticeAccess(t, testDoorCode, testAccessInstructions)
	status, counts = accessUpdateCall(t, http.MethodPost, adminToken)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, domain.AccessUpdateRecipients{Recipients: 2, WithoutEmail: 0}, counts)

	emails := waitForAccessUpdates(t, 2)
	require.Len(t, emails, 2, "one email per upcoming confirmed booking with an email")
	byBooking := map[string]notificationDomain.BookingAccessUpdateRequest{}
	for _, e := range emails {
		byBooking[e.BookingID] = e
	}
	require.Contains(t, byBooking, french.ID.String())
	require.Contains(t, byBooking, english.ID.String())
	assert.Equal(t, "fr.access@example.com", byBooking[french.ID.String()].ToEmail)
	assert.Equal(t, notificationDomain.LocaleFR, byBooking[french.ID.String()].Locale)
	assert.Equal(t, "en.access@example.com", byBooking[english.ID.String()].ToEmail)
	assert.Equal(t, notificationDomain.LocaleEN, byBooking[english.ID.String()].Locale)
	for _, e := range emails {
		assert.Equal(t, testDoorCode, e.DoorCode)
		assert.Equal(t, testAccessInstructions, e.AccessInstructions)
		assert.True(t, strings.HasPrefix(e.ManageURL, "https://app.leviosa.test/"), "manage link %q", e.ManageURL)
	}

	entry, ok := th.GetLatestAuditEntry(t, ctx, testPool, "practice_access", "")
	require.True(t, ok, "expected an audit entry for the access update")
	assert.Equal(t, "booking.access_update_sent", entry.Action)
	assert.Equal(t, "administrator", entry.ActorRole)
	assert.EqualValues(t, 2, entry.Payload["recipients"])
	assert.EqualValues(t, 0, entry.Payload["without_email"])
	assert.NotContains(t, fmt.Sprint(entry.Payload), testDoorCode, "the audit entry never carries the code")
}

// Both access-update endpoints refuse anonymous and non-administrator
// requests.
func TestAccessUpdate_AdminOnly(t *testing.T) {
	ctx := context.Background()
	defer tu.ClearAuthData(t, ctx, authCtx)

	standardToken := tu.SetupStandardUser(t, ctx, authCtx)
	partnerToken := tu.SetupPartnerUser(t, ctx, authCtx)

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		status, _ := accessUpdateCall(t, method, "")
		assert.Equal(t, http.StatusUnauthorized, status, "%s anonymous", method)
		status, _ = accessUpdateCall(t, method, standardToken)
		assert.Equal(t, http.StatusForbidden, status, "%s standard user", method)
		status, _ = accessUpdateCall(t, method, partnerToken)
		assert.Equal(t, http.StatusForbidden, status, "%s partner", method)
	}
}
