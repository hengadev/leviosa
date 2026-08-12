package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"
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

// setupAuditBooking creates a fresh building/room/product/availability and
// books it as a client, returning the booking ID and both client/partner
// access tokens. Mirrors TestCancelBooking's setup helper.
func setupAuditBooking(t *testing.T, ctx context.Context, client *http.Client) (uuid.UUID, string, string) {
	t.Helper()

	tbooking.ClearBookingsTable(t, ctx, testPool)
	tavail.ClearAvailabilityTable(t, ctx, testPool)
	tr.ClearRoomsTable(t, ctx, testPool)
	tb.ClearBuildingsTable(t, ctx, testPool)
	tcatalog.ClearProductsTable(t, ctx, testPool)
	th.ClearAuditEntriesTable(t, ctx, testPool)

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

	partnerToken, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "audit-partner@test.com", room.ID, testPool, authCtx.Redis, crypto)
	clientToken, clientID := tsetup.SetupStandardUser(t, ctx, "audit-client@test.com", room.ID, testPool, authCtx.Redis, crypto)

	startTime := time.Now().Add(72 * time.Hour).Truncate(time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, endTime, 5000)
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

	reqCreate, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

	respCreate, err := client.Do(reqCreate)
	require.NoError(t, err)
	defer respCreate.Body.Close()
	require.Equal(t, http.StatusCreated, respCreate.StatusCode)

	var createdBooking domain.BookingResponse
	require.NoError(t, json.NewDecoder(respCreate.Body).Decode(&createdBooking))

	return createdBooking.ID, clientToken, partnerToken
}

// TestRefundBooking_WritesAuditEntry confirms issue 019's acceptance
// criterion for the booking refund use case: a successful refund writes an
// audit entry with the correct actor, action, target, and payload.
func TestRefundBooking_WritesAuditEntry(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	defer tu.ClearAuthData(t, ctx, authCtx)

	bookingID, clientToken, _ := setupAuditBooking(t, ctx, client)

	// Mark payment as paid: retrieve the payment intent CreateBooking already
	// attached, confirm it via the mock payment service directly (there's no
	// HTTP path to move a Stripe-mocked intent to "succeeded" from
	// requires_payment_method), then drive ProcessPayment via HTTP so the
	// booking record itself reflects paid status.
	dbBooking, err := bookingRepo.GetByID(ctx, bookingID)
	require.NoError(t, err)
	decrypted, err := domain.DecryptBookingEncx(ctx, crypto, dbBooking)
	require.NoError(t, err)
	require.NotNil(t, decrypted.PaymentIntentID, "CreateBooking should have attached a payment intent for a priced product")

	require.NoError(t, paymentService.ConfirmPaymentIntent(ctx, *decrypted.PaymentIntentID))

	payReqBody, err := json.Marshal(domain.ProcessPaymentRequest{PaymentIntentID: *decrypted.PaymentIntentID})
	require.NoError(t, err)
	payReq, err := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/payment", bytes.NewReader(payReqBody))
	require.NoError(t, err)
	payReq.Header.Set("Content-Type", "application/json")
	payReq.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})
	payResp, err := client.Do(payReq)
	require.NoError(t, err)
	defer payResp.Body.Close()
	require.Equal(t, http.StatusOK, payResp.StatusCode)

	adminToken := tu.SetupAdminUser(t, ctx, authCtx)

	refundReq, err := http.NewRequest("POST", testServerURL+"/bookings/"+bookingID.String()+"/refund", nil)
	require.NoError(t, err)
	refundReq.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: adminToken})

	refundResp, err := client.Do(refundReq)
	require.NoError(t, err)
	defer refundResp.Body.Close()
	require.Equal(t, http.StatusOK, refundResp.StatusCode)

	entry, ok := th.GetLatestAuditEntry(t, ctx, testPool, "booking", bookingID.String())
	require.True(t, ok, "expected an audit entry for the refund")

	assert.Equal(t, "booking.refunded", entry.Action)
	assert.Equal(t, "administrator", entry.ActorRole)
	assert.NotEqual(t, uuid.Nil, entry.ActorID)
	assert.NotEmpty(t, entry.Payload["refund_id"])
	assert.Equal(t, *decrypted.PaymentIntentID, entry.Payload["payment_intent_id"])
}

// TestAdminOverrideCancelBooking_WritesAuditEntry confirms issue 019's
// acceptance criterion for the booking admin-override use case. Issue 019
// named "booking refund and admin-override" as the two booking use cases to
// wire audit into; no admin-override capability existed yet, so this issue
// adds AdminOverrideCancelBooking (an admin-only cancellation that bypasses
// the cancellation-window policy) alongside the audit wiring — see
// POST /admin/bookings/{id}/override-cancel.
func TestAdminOverrideCancelBooking_WritesAuditEntry(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	defer tu.ClearAuthData(t, ctx, authCtx)

	bookingID, _, _ := setupAuditBooking(t, ctx, client)
	adminToken := tu.SetupAdminUser(t, ctx, authCtx)

	overrideReq, err := http.NewRequest("POST", testServerURL+"/admin/bookings/"+bookingID.String()+"/override-cancel",
		bytes.NewReader(mustJSON(t, domain.CancelBookingRequest{Reason: "Customer emergency, waiving cancellation window"})))
	require.NoError(t, err)
	overrideReq.Header.Set("Content-Type", "application/json")
	overrideReq.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: adminToken})

	overrideResp, err := client.Do(overrideReq)
	require.NoError(t, err)
	defer overrideResp.Body.Close()
	require.Equal(t, http.StatusOK, overrideResp.StatusCode)

	var cancelled domain.BookingResponse
	require.NoError(t, json.NewDecoder(overrideResp.Body).Decode(&cancelled))
	assert.Equal(t, domain.BookingStatusCancelled, cancelled.Status)

	entry, ok := th.GetLatestAuditEntry(t, ctx, testPool, "booking", bookingID.String())
	require.True(t, ok, "expected an audit entry for the admin override")

	assert.Equal(t, "booking.admin_override", entry.Action)
	assert.Equal(t, "administrator", entry.ActorRole)
	assert.NotEqual(t, uuid.Nil, entry.ActorID)
	assert.Equal(t, "Customer emergency, waiving cancellation window", entry.Payload["reason"])
	assert.Equal(t, true, entry.Payload["bypassed_cancellation_policy"])
}

// TestAdminOverrideCancelBooking_RequiresAdmin confirms the override route is
// admin-gated, not reachable by a standard client.
func TestAdminOverrideCancelBooking_RequiresAdmin(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}
	defer tu.ClearAuthData(t, ctx, authCtx)

	bookingID, clientToken, _ := setupAuditBooking(t, ctx, client)

	overrideReq, err := http.NewRequest("POST", testServerURL+"/admin/bookings/"+bookingID.String()+"/override-cancel",
		bytes.NewReader(mustJSON(t, domain.CancelBookingRequest{Reason: "not an admin"})))
	require.NoError(t, err)
	overrideReq.Header.Set("Content-Type", "application/json")
	overrideReq.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: clientToken})

	overrideResp, err := client.Do(overrideReq)
	require.NoError(t, err)
	defer overrideResp.Body.Close()
	assert.Equal(t, http.StatusForbidden, overrideResp.StatusCode)
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
