package booking

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	tsetup "github.com/Leviosa-care/leviosa/backend/test/helpers/booking"
	tavail "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/availability"
	tbooking "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/booking"
	tb "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/building"
	tr "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/room"
	tcatalog "github.com/Leviosa-care/leviosa/backend/test/helpers/catalog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A booking confirmed by the payment webhook sends exactly one client email,
// in the locale the client booked in, one partner notice, and no SMS
// (issue 016). The booking-confirmed event goes through the real
// BookingNotificationAdapter down to capturing fakes at the email and SMS
// ports.
func TestBookingConfirmedEmails(t *testing.T) {
	ctx := context.Background()

	tbooking.ClearBookingsTable(t, ctx, testPool)
	tavail.ClearAvailabilityTable(t, ctx, testPool)
	tr.ClearRoomsTable(t, ctx, testPool)
	tb.ClearBuildingsTable(t, ctx, testPool)
	tcatalog.ClearProductsTable(t, ctx, testPool)
	defer tu.ClearAuthData(t, ctx, authCtx)

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

	_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, "partner-emails@example.com", room.ID, testPool, authCtx.Redis, crypto)

	startTime := time.Now().Add(48 * time.Hour).Truncate(time.Hour)
	availability := tavail.NewTestAvailability(t, partnerID, room.ID, startTime, startTime.Add(2*time.Hour), 0)
	availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
	require.NoError(t, err)
	tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

	// A guest books in English.
	booking, err := service.CreateBooking(ctx, availability.ID, nil, products[0].ID, startTime,
		"Jane", "Doe", "jane.doe@example.com", "+33612345678", domain.BookingLocaleEN)
	require.NoError(t, err)
	require.Equal(t, domain.BookingStatusAwaitingPayment, booking.Status)
	require.NotEmpty(t, booking.Token)

	// Follow booking-confirmed events down to the real adapter.
	notificationEmailSpy.Reset()
	smsBefore := notificationSMSSpy.Count()
	stop := notificationSpy.ForwardConfirmationsTo(realNotificationAdapter)
	defer stop()

	// Stripe confirms the payment, twice; the PaymentIntent charged less than
	// the total price (a deposit), so a balance remains due at the practice.
	const depositCents = 1500
	event := paymentSucceededEvent(t, ctx, booking.ID)
	event.Amount = depositCents
	require.NoError(t, service.HandlePaymentWebhook(ctx, event))
	require.NoError(t, service.HandlePaymentWebhook(ctx, event))

	require.Eventually(t, func() bool {
		clients, partners := notificationEmailSpy.Snapshot()
		return len(clients) >= 1 && len(partners) >= 1
	}, 5*time.Second, 20*time.Millisecond, "expected the client confirmation and the partner notice")
	// Give a duplicate a chance to show up before counting.
	time.Sleep(200 * time.Millisecond)
	clients, partners := notificationEmailSpy.Snapshot()

	require.Len(t, clients, 1, "exactly one client confirmation")
	client := clients[0]
	assert.Equal(t, "jane.doe@example.com", client.ToEmail)
	assert.Equal(t, "Jane", client.ToFirstName)
	assert.Equal(t, notificationDomain.LocaleEN, client.Locale, "the client email follows the booking's locale")
	assert.Equal(t, booking.ID.String(), client.BookingID)
	assert.Equal(t, stubPracticeAddress, client.Address, "the practice address comes from the company settings")
	assert.True(t, client.StartTime.Equal(booking.SlotStartTime))
	assert.True(t, client.EndTime.Equal(booking.SlotEndTime))
	assert.Equal(t, depositCents, client.AmountPaidOnlineCents)
	assert.Equal(t, booking.TotalPriceCents-depositCents, client.BalanceDueCents)
	assert.NotEmpty(t, client.ProductName)

	// The manage link opens the booking token page for this booking.
	manage, err := url.Parse(client.ManageURL)
	require.NoError(t, err)
	assert.Equal(t, "https://app.leviosa.test", manage.Scheme+"://"+manage.Host)
	assert.Equal(t, "/bookings", manage.Path)
	assert.Equal(t, booking.Token, manage.Query().Get("token"))

	require.Len(t, partners, 1, "exactly one partner notice")
	partner := partners[0]
	assert.Equal(t, "partner-emails@example.com", partner.ToEmail)
	assert.Equal(t, "Jane Doe", partner.ClientName)
	assert.Equal(t, "jane.doe@example.com", partner.ClientEmail)
	assert.Equal(t, "+33612345678", partner.ClientPhone)
	assert.Equal(t, booking.ID.String(), partner.BookingID)

	assert.Equal(t, smsBefore, notificationSMSSpy.Count(), "booking notifications send no SMS")

	// The composed client email is English and carries the .ics.
	composed := notificationDomain.ComposeBookingConfirmation(client, time.Now())
	require.Len(t, composed.Attachments, 1)
	assert.Equal(t, "appointment.ics", composed.Attachments[0].Filename)
	assert.Contains(t, string(composed.Attachments[0].Data),
		"DTSTART:"+booking.SlotStartTime.UTC().Format("20060102T150405Z"))
}
