package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	tsetup "github.com/Leviosa-care/leviosa/backend/test/helpers/booking"
	tavail "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/availability"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Availability exceptions (docs/issues/012-availability-exceptions.md): the
// partner closes a date or a range on a recurring availability; the practice
// page offers no slot on it and booking creation refuses it; the bookings
// already there are counted and left as they are; only the series' partner
// sees and changes its exceptions.
//
// make test-func TEST_NAME=TestAvailabilityExceptions TEST_PATH=test/integration/booking/booking/availability_exceptions_test.go

func partnerToken(t *testing.T, ctx context.Context, f slotFixture) string {
	t.Helper()
	return tsetup.CreateSessionForUser(t, ctx, f.partnerUserID, identity.Partner, redisClient, crypto)
}

func exceptionsURL(availabilityID uuid.UUID) string {
	return fmt.Sprintf("%s/availabilities/%s/exceptions", testServerURL, availabilityID)
}

func doAsPartner(t *testing.T, client *http.Client, token, method, url string, body any) *http.Response {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: token})
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

// closeDates creates an exception from start to end (practice dates) and
// returns the response status and body.
func closeDates(t *testing.T, client *http.Client, token string, availabilityID uuid.UUID, start, end time.Time) (int, domain.CreateAvailabilityExceptionResponse) {
	t.Helper()
	resp := doAsPartner(t, client, token, http.MethodPost, exceptionsURL(availabilityID), map[string]string{
		"start_date": start.In(domain.PracticeLocation).Format(time.DateOnly),
		"end_date":   end.In(domain.PracticeLocation).Format(time.DateOnly),
	})
	defer resp.Body.Close()
	var body domain.CreateAvailabilityExceptionResponse
	if resp.StatusCode == http.StatusCreated {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	}
	return resp.StatusCode, body
}

func listExceptions(t *testing.T, client *http.Client, token string, availabilityID uuid.UUID) (int, []domain.AvailabilityExceptionResponse) {
	t.Helper()
	resp := doAsPartner(t, client, token, http.MethodGet, exceptionsURL(availabilityID), nil)
	defer resp.Body.Close()
	var body []domain.AvailabilityExceptionResponse
	if resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	}
	return resp.StatusCode, body
}

func deleteException(t *testing.T, client *http.Client, token string, availabilityID, exceptionID uuid.UUID) int {
	t.Helper()
	resp := doAsPartner(t, client, token, http.MethodDelete, fmt.Sprintf("%s/%s", exceptionsURL(availabilityID), exceptionID), nil)
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestAvailabilityExceptions(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("should offer no slot on a closed date and refuse a booking on it", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		availability := createWeeklyAvailability(t, ctx, f)
		token := partnerToken(t, ctx, f)

		status, created := closeDates(t, client, token, availability.ID, f.onDay(7, 0, 0), f.onDay(7, 0, 0))
		require.Equal(t, http.StatusCreated, status)
		assert.Equal(t, 0, created.AffectedBookings)
		assert.Equal(t, f.onDay(7, 0, 0).Format(time.DateOnly), created.Exception.StartDate)
		assert.Equal(t, created.Exception.StartDate, created.Exception.EndDate)

		assert.Empty(t, slotsOn(t, client, f, f.onDay(7, 0, 0)), "the closed date offers no slot")
		assertSameInstants(t, []time.Time{f.onDay(0, 9, 0), f.onDay(0, 10, 15), f.onDay(0, 11, 30)},
			slotsOn(t, client, f, f.onDay(0, 0, 0)))
		assertSameInstants(t, []time.Time{f.onDay(14, 9, 0), f.onDay(14, 10, 15), f.onDay(14, 11, 30)},
			slotsOn(t, client, f, f.onDay(14, 0, 0)))

		assert.Equal(t, http.StatusBadRequest, postBooking(t, client, availability.ID, f.productID, f.onDay(7, 10, 15)),
			"a start on a closed date is not a slot offered")
		assert.Equal(t, http.StatusCreated, postBooking(t, client, availability.ID, f.productID, f.onDay(14, 10, 15)))
	})

	t.Run("should close every occurrence of a date range, both ends included", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		availability := createWeeklyAvailability(t, ctx, f)
		token := partnerToken(t, ctx, f)

		status, _ := closeDates(t, client, token, availability.ID, f.onDay(7, 0, 0), f.onDay(14, 0, 0))
		require.Equal(t, http.StatusCreated, status)

		from := f.onDay(0, 0, 0).Format(time.DateOnly)
		to := f.onDay(21, 0, 0).Format(time.DateOnly)
		status, body := getFreeSlots(t, client, f.partnerID, f.productID, from, to)
		require.Equal(t, http.StatusOK, status)
		assertSameInstants(t, []time.Time{
			f.onDay(0, 9, 0), f.onDay(0, 10, 15), f.onDay(0, 11, 30),
			f.onDay(21, 9, 0), f.onDay(21, 10, 15), f.onDay(21, 11, 30),
		}, slotStarts(body))
		assert.Equal(t, http.StatusBadRequest, postBooking(t, client, availability.ID, f.productID, f.onDay(14, 9, 0)))
	})

	t.Run("should report the bookings on the closed dates and leave them confirmed", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		availability := createWeeklyAvailability(t, ctx, f)
		token := partnerToken(t, ctx, f)

		book := func(start time.Time) *domain.Booking {
			t.Helper()
			booked, err := service.CreateBooking(ctx, availability.ID, &f.clientID, f.productID, start, "", "", "", "", "")
			require.NoError(t, err)
			return booked
		}
		confirmed := book(f.onDay(7, 9, 0))
		ok, err := bookingRepo.ConfirmAwaitingPayment(ctx, confirmed.ID)
		require.NoError(t, err)
		require.True(t, ok)
		held := book(f.onDay(7, 11, 30)) // awaiting payment, hold running: counted
		lapsed := book(f.onDay(7, 10, 15))
		_, err = testPool.Exec(ctx, `UPDATE booking.bookings SET hold_expires_at = NOW() - INTERVAL '1 minute' WHERE id = $1`, lapsed.ID)
		require.NoError(t, err) // hold elapsed: not counted
		otherWeek := book(f.onDay(14, 9, 0))
		_, err = bookingRepo.ConfirmAwaitingPayment(ctx, otherWeek.ID)
		require.NoError(t, err) // outside the closed date: not counted

		before := map[uuid.UUID]bookingSnapshot{}
		for _, id := range []uuid.UUID{confirmed.ID, held.ID, lapsed.ID, otherWeek.ID} {
			before[id] = snapshotBooking(t, ctx, id)
		}

		status, created := closeDates(t, client, token, availability.ID, f.onDay(7, 0, 0), f.onDay(7, 0, 0))
		require.Equal(t, http.StatusCreated, status)
		assert.Equal(t, 2, created.AffectedBookings, "the confirmed booking and the running hold")

		for id, snapshot := range before {
			assert.Equal(t, snapshot, snapshotBooking(t, ctx, id), "closing a date never changes a booking")
		}
		assert.Equal(t, domain.BookingStatusConfirmed, snapshotBooking(t, ctx, confirmed.ID).Status)
	})

	t.Run("should list and reopen exceptions for the series' partner only", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		availability := createWeeklyAvailability(t, ctx, f)
		token := partnerToken(t, ctx, f)
		otherToken, _ := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-other@example.com", f.roomID, testPool, authCtx.Redis, crypto)

		status, created := closeDates(t, client, token, availability.ID, f.onDay(7, 0, 0), f.onDay(7, 0, 0))
		require.Equal(t, http.StatusCreated, status)
		exceptionID := created.Exception.ID

		// Another partner can neither see, add nor remove them.
		status, _ = listExceptions(t, client, otherToken, availability.ID)
		assert.Equal(t, http.StatusNotFound, status)
		status, _ = closeDates(t, client, otherToken, availability.ID, f.onDay(14, 0, 0), f.onDay(14, 0, 0))
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, http.StatusNotFound, deleteException(t, client, otherToken, availability.ID, exceptionID))

		status, listed := listExceptions(t, client, token, availability.ID)
		require.Equal(t, http.StatusOK, status)
		require.Len(t, listed, 1, "the other partner added nothing")
		assert.Equal(t, exceptionID, listed[0].ID)
		assert.Empty(t, slotsOn(t, client, f, f.onDay(7, 0, 0)), "the other partner removed nothing")

		// The series' partner reopens the date.
		assert.Equal(t, http.StatusNoContent, deleteException(t, client, token, availability.ID, exceptionID))
		assert.Equal(t, http.StatusNotFound, deleteException(t, client, token, availability.ID, exceptionID))
		status, listed = listExceptions(t, client, token, availability.ID)
		require.Equal(t, http.StatusOK, status)
		assert.Empty(t, listed)
		assertSameInstants(t, []time.Time{f.onDay(7, 9, 0), f.onDay(7, 10, 15), f.onDay(7, 11, 30)},
			slotsOn(t, client, f, f.onDay(7, 0, 0)))
	})

	t.Run("should refuse an inverted range and a one-off availability", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupSlotPartner(t, ctx)
		availability := createWeeklyAvailability(t, ctx, f)
		token := partnerToken(t, ctx, f)

		status, _ := closeDates(t, client, token, availability.ID, f.onDay(8, 0, 0), f.onDay(7, 0, 0))
		assert.Equal(t, http.StatusBadRequest, status)

		oneOff := tavail.NewTestAvailability(t, f.partnerUserID, f.roomID, f.onDay(1, 9, 0), f.onDay(1, 13, 0), 0)
		oneOffEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, oneOff)
		require.NoError(t, err)
		tavail.InsertAvailabilityEncx(t, ctx, oneOffEncx, testPool)
		status, _ = closeDates(t, client, token, oneOff.ID, f.onDay(1, 0, 0), f.onDay(1, 0, 0))
		assert.Equal(t, http.StatusBadRequest, status)

		status, listed := listExceptions(t, client, token, availability.ID)
		require.Equal(t, http.StatusOK, status)
		assert.Empty(t, listed)
	})
}
