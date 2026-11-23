package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	settingsDomain "github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestPractitionerBooking TEST_PATH=test/integration/booking/booking/practitioner_booking_test.go

// The practitioner books a client who called (issue 041, ADR-0021): the
// slot engine's rules without the minimum notice and the horizon, confirmed
// with nothing paid online, no client message, and the same agenda lock as
// an online booking.

type practitionerBookingBody struct {
	ID                   uuid.UUID            `json:"id"`
	Status               domain.BookingStatus `json:"status"`
	Origin               domain.BookingOrigin `json:"origin"`
	SlotStartTime        time.Time            `json:"slot_start_time"`
	TotalPriceCents      int                  `json:"total_price_cents"`
	AmountDueOnSiteCents int                  `json:"amount_due_on_site_cents"`
}

func postPractitionerBooking(t *testing.T, client *http.Client, token string, body map[string]any) (int, practitionerBookingBody) {
	t.Helper()
	payload, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, testServerURL+"/partners/bookings", bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: token})
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var created practitionerBookingBody
	if resp.StatusCode == http.StatusCreated {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	}
	return resp.StatusCode, created
}

func practitionerBookingRequest(f slotFixture, start time.Time) map[string]any {
	return map[string]any{
		"availability_id":  f.availabilityID.String(),
		"product_id":       f.productID.String(),
		"slot_start_time":  start.Format(time.RFC3339),
		"guest_first_name": "Marie",
		"guest_last_name":  "Dupont",
		"guest_phone":      "0612345678",
		"guest_email":      "marie@example.com",
	}
}

func getPractitionerSlots(t *testing.T, client *http.Client, token string, partnerUserID, productID uuid.UUID, day string) (int, domain.FreeSlotsResponse) {
	t.Helper()
	url := fmt.Sprintf("%s/partners/booking-slots/%s?product_id=%s&from=%s&to=%s", testServerURL, partnerUserID, productID, day, day)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: token})
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var body domain.FreeSlotsResponse
	if resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	}
	return resp.StatusCode, body
}

// setupPractitionerFixture is setupSlotFixture with the product among the
// practitioner's own (their profile's products).
func setupPractitionerFixture(t *testing.T, ctx context.Context) slotFixture {
	t.Helper()
	f := setupSlotFixture(t, ctx)
	offerProducts(t, ctx, f.partnerUserID, f.productID)
	return f
}

// offerProducts sets the products the partner's profile offers.
func offerProducts(t *testing.T, ctx context.Context, partnerUserID uuid.UUID, productIDs ...uuid.UUID) {
	t.Helper()
	_, err := testPool.Exec(ctx, "UPDATE auth.partners SET product_ids = $2 WHERE user_id = $1", partnerUserID, productIDs)
	require.NoError(t, err)
}

func (m *MockPaymentService) refundCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.refunds)
}

func TestPractitionerBooking(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	defer setBookingPolicy(t, ctx, suiteMinimumNoticeHours, settingsDomain.DefaultBookingHorizonWeeks)

	t.Run("should book a client who called within the minimum notice, telling the client nothing", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupPractitionerFixture(t, ctx)
		// A week of notice: the fixture's day, three days out, is closed to
		// the practice page.
		setBookingPolicy(t, ctx, 168, 8)
		defer setBookingPolicy(t, ctx, suiteMinimumNoticeHours, settingsDomain.DefaultBookingHorizonWeeks)
		day := f.day.Format(time.DateOnly)

		status, public := getFreeSlots(t, client, f.partnerID, f.productID, day, day)
		require.Equal(t, http.StatusOK, status)
		assert.Empty(t, public.Slots, "the public practice page applies the minimum notice")

		status, slots := getPractitionerSlots(t, client, f.partnerToken, f.partnerUserID, f.productID, day)
		require.Equal(t, http.StatusOK, status)
		assertSameInstants(t, []time.Time{f.at(9, 0), f.at(10, 15), f.at(11, 30)}, slotStarts(slots))

		status, created := postPractitionerBooking(t, client, f.partnerToken, practitionerBookingRequest(f, f.at(10, 15)))
		require.Equal(t, http.StatusCreated, status)
		assert.Equal(t, domain.BookingStatusConfirmed, created.Status)
		assert.Equal(t, domain.BookingOriginPractitioner, created.Origin)
		assert.True(t, created.SlotStartTime.Equal(f.at(10, 15)))
		assert.Positive(t, created.TotalPriceCents)
		assert.Equal(t, created.TotalPriceCents, created.AmountDueOnSiteCents, "the full price is due at the practice")

		stored := storedOnSiteBooking(t, ctx, created.ID)
		assert.Equal(t, domain.BookingOriginPractitioner, stored.Origin)
		require.NotNil(t, stored.AmountPaidOnlineCents)
		assert.Equal(t, 0, *stored.AmountPaidOnlineCents)
		assert.Nil(t, stored.PaymentIntentID)
		assert.Nil(t, stored.HoldExpiresAt)
		assert.Empty(t, stored.Token)
		assert.Equal(t, "0612345678", stored.GuestPhone)
		assert.Equal(t, 0, notificationSpy.CountFor(func(s *SpyNotificationService) []ports.BookingNotificationData { return s.BookingConfirmations }, created.ID),
			"no confirmation")

		// Aujourd'hui and Réservations show it with the full price due.
		partnerBookings, err := service.GetPartnerBookingsEnriched(ctx, f.partnerUserID, ports.BookingFilter{})
		require.NoError(t, err)
		require.Len(t, partnerBookings, 1)
		assert.Equal(t, created.TotalPriceCents, partnerBookings[0].AmountDueOnSiteCents)
		assert.Equal(t, domain.BookingOriginPractitioner, partnerBookings[0].Origin)

		// The client's 24-hour reminder query leaves it out.
		due, err := bookingRepo.FindBookingsDueForReminder(ctx)
		require.NoError(t, err)
		for _, b := range due {
			assert.NotEqual(t, created.ID, b.ID, "no reminder to a practitioner booking's client")
		}

		// The slot is taken now.
		status, slots = getPractitionerSlots(t, client, f.partnerToken, f.partnerUserID, f.productID, day)
		require.Equal(t, http.StatusOK, status)
		assertSameInstants(t, []time.Time{f.at(9, 0), f.at(11, 30)}, slotStarts(slots))
		status, _ = postPractitionerBooking(t, client, f.partnerToken, practitionerBookingRequest(f, f.at(10, 15)))
		assert.Equal(t, http.StatusConflict, status)

		// Cancelled through the partner cancellation: no refund, no email,
		// the slot is free again.
		refunds := paymentService.(*MockPaymentService).refundCount()
		req, err := http.NewRequest(http.MethodPost, testServerURL+"/partners/bookings/"+created.ID.String()+"/cancel", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: f.partnerToken})
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		cancelled := storedOnSiteBooking(t, ctx, created.ID)
		assert.Equal(t, domain.BookingStatusCancelled, cancelled.Status)
		require.NotNil(t, cancelled.RefundedAmountCents)
		assert.Equal(t, 0, *cancelled.RefundedAmountCents)
		assert.Equal(t, refunds, paymentService.(*MockPaymentService).refundCount(), "nothing refunded")
		assert.Equal(t, 0, notificationSpy.CountFor(func(s *SpyNotificationService) []ports.BookingNotificationData { return s.BookingCancellations }, created.ID),
			"no cancellation email")

		status, slots = getPractitionerSlots(t, client, f.partnerToken, f.partnerUserID, f.productID, day)
		require.Equal(t, http.StatusOK, status)
		assertSameInstants(t, []time.Time{f.at(9, 0), f.at(10, 15), f.at(11, 30)}, slotStarts(slots))
	})

	t.Run("should refuse a slot outside the availability or off the grid", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupPractitionerFixture(t, ctx)

		for _, start := range []time.Time{f.at(8, 0), f.at(9, 30), f.at(12, 45), f.at(14, 0)} {
			status, _ := postPractitionerBooking(t, client, f.partnerToken, practitionerBookingRequest(f, start))
			assert.Equal(t, http.StatusBadRequest, status, "start %s", start)
		}
	})

	t.Run("should require the client's phone", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupPractitionerFixture(t, ctx)

		body := practitionerBookingRequest(f, f.at(9, 0))
		delete(body, "guest_phone")
		status, _ := postPractitionerBooking(t, client, f.partnerToken, body)
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("should let only the practitioner or an administrator book", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupPractitionerFixture(t, ctx)
		day := f.day.Format(time.DateOnly)

		otherPartner := tu.SetupPartnerUser(t, ctx, authCtx)
		status, _ := postPractitionerBooking(t, client, otherPartner, practitionerBookingRequest(f, f.at(9, 0)))
		assert.Equal(t, http.StatusForbidden, status)
		status, _ = getPractitionerSlots(t, client, otherPartner, f.partnerUserID, f.productID, day)
		assert.Equal(t, http.StatusForbidden, status)

		admin := tu.SetupAdminUser(t, ctx, authCtx)
		status, slots := getPractitionerSlots(t, client, admin, f.partnerUserID, f.productID, day)
		require.Equal(t, http.StatusOK, status)
		assert.Len(t, slots.Slots, 3)
		status, created := postPractitionerBooking(t, client, admin, practitionerBookingRequest(f, f.at(9, 0)))
		require.Equal(t, http.StatusCreated, status)
		assert.Equal(t, domain.BookingOriginPractitioner, created.Origin)
	})

	t.Run("should list the practitioner's bookable products", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupPractitionerFixture(t, ctx)

		req, err := http.NewRequest(http.MethodGet, testServerURL+"/partners/booking-products/"+f.partnerUserID.String(), nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: f.partnerToken})
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var products []domain.PractitionerProduct
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&products))
		var ids []uuid.UUID
		for _, p := range products {
			ids = append(ids, p.ID)
			assert.Positive(t, p.PriceCents)
		}
		assert.Equal(t, []uuid.UUID{f.productID}, ids, "the products of the practitioner's profile")
	})

	t.Run("should refuse a product the practitioner does not offer", func(t *testing.T) {
		defer tu.ClearAuthData(t, ctx, authCtx)
		f := setupPractitionerFixture(t, ctx)
		offerProducts(t, ctx, f.partnerUserID, uuid.New())

		status, _ := postPractitionerBooking(t, client, f.partnerToken, practitionerBookingRequest(f, f.at(9, 0)))
		assert.Equal(t, http.StatusConflict, status)
	})
}

// make test-func TEST_NAME=TestPractitionerAndOnlineBookingOfTheSameSlot TEST_PATH=test/integration/booking/booking/practitioner_booking_test.go

// A practitioner booking and an online booking of the same slot, at once:
// the partner's agenda lock lets exactly one of them have it.
func TestPractitionerAndOnlineBookingOfTheSameSlot(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	defer tu.ClearAuthData(t, ctx, authCtx)
	f := setupPractitionerFixture(t, ctx)

	for _, start := range []time.Time{f.at(9, 0), f.at(10, 15), f.at(11, 30)} {
		var wg sync.WaitGroup
		ready := make(chan struct{})
		codes := make([]int, 2)
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-ready
			codes[0], _ = postPractitionerBooking(t, client, f.partnerToken, practitionerBookingRequest(f, start))
		}()
		go func() {
			defer wg.Done()
			<-ready
			codes[1] = postBooking(t, client, f.availabilityID, f.productID, start)
		}()
		close(ready)
		wg.Wait()

		sort.Ints(codes)
		assert.Equal(t, []int{http.StatusCreated, http.StatusConflict}, codes, "slot %s", start)
	}

	// One booking per slot holds it.
	bookingsEncx, err := bookingRepo.GetByPartnerID(ctx, f.partnerUserID, ports.BookingFilter{})
	require.NoError(t, err)
	perStart := map[time.Time]int{}
	for _, bx := range bookingsEncx {
		b, err := domain.DecryptBookingEncx(ctx, crypto, bx)
		require.NoError(t, err)
		if b.OccupiesTime(time.Now()) {
			perStart[b.SlotStartTime.UTC()]++
		}
	}
	assert.Len(t, perStart, 3)
	for start, n := range perStart {
		assert.Equal(t, 1, n, "slot %s", start)
	}
}
