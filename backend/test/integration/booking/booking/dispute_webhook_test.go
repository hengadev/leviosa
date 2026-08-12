package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	tsetup "github.com/Leviosa-care/leviosa/backend/test/helpers/booking"
	tavail "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/availability"
	tbooking "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/booking"
	tb "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/building"
	tr "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/room"
	tcatalog "github.com/Leviosa-care/leviosa/backend/test/helpers/catalog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupPaidBookingForDispute creates a confirmed, paid booking (via HTTP creation +
// a payment_intent.succeeded webhook, mirroring TestPaymentConfirmationNotification)
// and returns its ID and the Stripe payment intent ID it is linked to.
func setupPaidBookingForDispute(t *testing.T, ctx context.Context, httpClient *http.Client, emailSuffix string) (bookingID uuid.UUID, paymentIntentID string) {
	t.Helper()

	tbooking.ClearBookingsTable(t, ctx, testPool)
	tavail.ClearAvailabilityTable(t, ctx, testPool)
	tr.ClearRoomsTable(t, ctx, testPool)
	tb.ClearBuildingsTable(t, ctx, testPool)
	tcatalog.ClearProductsTable(t, ctx, testPool)

	categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
	products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)

	building := tb.NewTestBuilding(t)
	buildingEncx, err := domain.ProcessBuildingEncx(ctx, crypto, building)
	require.NoError(t, err)
	require.NoError(t, tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx))

	room := tr.NewTestRoomWithBuilding(t, building.ID)
	roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
	require.NoError(t, err)
	require.NoError(t, tr.InsertRoomEncx(t, ctx, testPool, roomEncx))

	_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-dispute"+emailSuffix+"@example.com", room.ID, testPool, authCtx.Redis, crypto)
	clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "client-dispute"+emailSuffix+"@example.com", room.ID, testPool, authCtx.Redis, crypto)

	startTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, endTime, 0)
	availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
	require.NoError(t, err)
	tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

	productID := products[0].ID
	slotStartTime := startTime.Add(30 * time.Minute).Truncate(10 * time.Minute)
	requestBody := map[string]interface{}{
		"availability_id": availability.ID.String(),
		"client_id":       clientID,
		"product_id":      productID.String(),
		"slot_start_time": slotStartTime.Format(time.RFC3339),
	}
	bodyBytes, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var booking domain.BookingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&booking))

	paymentIntentID = "pi_test_" + uuid.New().String()[:8]

	// Attach the payment intent ID to the booking (as CreatePaymentIntent would).
	bookingEncx, err := bookingRepo.GetByID(ctx, booking.ID)
	require.NoError(t, err)
	decryptedBooking, err := domain.DecryptBookingEncx(ctx, crypto, bookingEncx)
	require.NoError(t, err)
	decryptedBooking.SetPaymentIntentID(paymentIntentID)
	updatedEncx, err := domain.ProcessBookingEncx(ctx, crypto, decryptedBooking)
	require.NoError(t, err)
	require.NoError(t, bookingRepo.Update(ctx, updatedEncx))

	// Mark it paid via the standard payment_intent.succeeded webhook path.
	succeededEvent := &ports.WebhookEvent{
		ID:              "evt_test_" + uuid.New().String(),
		Type:            ports.WebhookEventPaymentIntentSucceeded,
		PaymentIntentID: paymentIntentID,
		Status:          "succeeded",
		Amount:          booking.TotalPriceCents,
		Currency:        booking.Currency,
		Metadata: map[string]string{
			"booking_id": booking.ID.String(),
		},
	}
	require.NoError(t, service.HandlePaymentWebhook(ctx, succeededEvent))

	return booking.ID, paymentIntentID
}

func TestDisputeWebhook_MarksBookingDisputedAndAlertsAdmin(t *testing.T) {
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 10 * time.Second}
	defer tu.ClearAuthData(t, ctx, authCtx)

	bookingID, paymentIntentID := setupPaidBookingForDispute(t, ctx, httpClient, "-created")

	// Sanity check: booking is paid before the dispute.
	encx, err := bookingRepo.GetByID(ctx, bookingID)
	require.NoError(t, err)
	b, err := domain.DecryptBookingEncx(ctx, crypto, encx)
	require.NoError(t, err)
	require.Equal(t, domain.PaymentStatusPaid, b.PaymentStatus)

	notificationSpy.DisputeAlerts = nil

	disputeEvent := &ports.WebhookEvent{
		ID:              "evt_test_" + uuid.New().String(),
		Type:            ports.WebhookEventChargeDisputeCreated,
		PaymentIntentID: paymentIntentID,
		Status:          "needs_response",
		Amount:          b.TotalPriceCents,
		Currency:        b.Currency,
		DisputeID:       "dp_test_" + uuid.New().String()[:8],
		DisputeReason:   "fraudulent",
		// Dispute events carry no booking_id metadata (see HandlePaymentWebhook doc) —
		// resolution happens via PaymentIntentID instead.
		Metadata: map[string]string{},
	}

	err = service.HandlePaymentWebhook(ctx, disputeEvent)
	require.NoError(t, err)

	// Assert: booking payment status is now disputed.
	encx, err = bookingRepo.GetByID(ctx, bookingID)
	require.NoError(t, err)
	b, err = domain.DecryptBookingEncx(ctx, crypto, encx)
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusDisputed, b.PaymentStatus, "booking payment status should be disputed")

	// Assert: no automatic cancellation happened.
	assert.Equal(t, domain.BookingStatusConfirmed, b.Status, "booking status must not change automatically on dispute")

	// Assert: admin was alerted, linked to this booking.
	require.Len(t, notificationSpy.DisputeAlerts, 1, "expected exactly one dispute alert notification")
	alert := notificationSpy.DisputeAlerts[0]
	assert.Equal(t, bookingID, alert.BookingID)
	assert.Equal(t, "fraudulent", alert.DisputeReason)

	// A duplicate delivery of the same event must not send a second alert or error out.
	notificationSpy.DisputeAlerts = nil
	err = service.HandlePaymentWebhook(ctx, disputeEvent)
	require.NoError(t, err)
	assert.Empty(t, notificationSpy.DisputeAlerts, "duplicate dispute.created delivery should not re-alert")
}

func TestDisputeWebhook_ClosedLostRevertsToRefunded(t *testing.T) {
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 10 * time.Second}
	defer tu.ClearAuthData(t, ctx, authCtx)

	bookingID, paymentIntentID := setupPaidBookingForDispute(t, ctx, httpClient, "-lost")

	createdEvent := &ports.WebhookEvent{
		ID:              "evt_test_" + uuid.New().String(),
		Type:            ports.WebhookEventChargeDisputeCreated,
		PaymentIntentID: paymentIntentID,
		Status:          "needs_response",
		DisputeID:       "dp_test_" + uuid.New().String()[:8],
		DisputeReason:   "fraudulent",
	}
	require.NoError(t, service.HandlePaymentWebhook(ctx, createdEvent))

	closedEvent := &ports.WebhookEvent{
		ID:              "evt_test_" + uuid.New().String(),
		Type:            ports.WebhookEventChargeDisputeClosed,
		PaymentIntentID: paymentIntentID,
		Status:          "lost",
		DisputeID:       createdEvent.DisputeID,
		DisputeReason:   "fraudulent",
	}
	err := service.HandlePaymentWebhook(ctx, closedEvent)
	require.NoError(t, err)

	encx, err := bookingRepo.GetByID(ctx, bookingID)
	require.NoError(t, err)
	b, err := domain.DecryptBookingEncx(ctx, crypto, encx)
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusRefunded, b.PaymentStatus, "a lost dispute should resolve to refunded")
}

func TestDisputeWebhook_ClosedWonRevertsToPaid(t *testing.T) {
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 10 * time.Second}
	defer tu.ClearAuthData(t, ctx, authCtx)

	bookingID, paymentIntentID := setupPaidBookingForDispute(t, ctx, httpClient, "-won")

	createdEvent := &ports.WebhookEvent{
		ID:              "evt_test_" + uuid.New().String(),
		Type:            ports.WebhookEventChargeDisputeCreated,
		PaymentIntentID: paymentIntentID,
		Status:          "needs_response",
		DisputeID:       "dp_test_" + uuid.New().String()[:8],
		DisputeReason:   "duplicate",
	}
	require.NoError(t, service.HandlePaymentWebhook(ctx, createdEvent))

	closedEvent := &ports.WebhookEvent{
		ID:              "evt_test_" + uuid.New().String(),
		Type:            ports.WebhookEventChargeDisputeClosed,
		PaymentIntentID: paymentIntentID,
		Status:          "won",
		DisputeID:       createdEvent.DisputeID,
		DisputeReason:   "duplicate",
	}
	err := service.HandlePaymentWebhook(ctx, closedEvent)
	require.NoError(t, err)

	encx, err := bookingRepo.GetByID(ctx, bookingID)
	require.NoError(t, err)
	b, err := domain.DecryptBookingEncx(ctx, crypto, encx)
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusPaid, b.PaymentStatus, "a won dispute should resolve back to paid")
}
