package booking

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"

	bookingService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/booking"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
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

// reminderFixture is a room with a partner, ready to take bookings.
type reminderFixture struct {
	partnerID uuid.UUID
	roomID    uuid.UUID
	productID uuid.UUID
}

func setupReminderFixture(t *testing.T, ctx context.Context, partnerEmail string) reminderFixture {
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

	_, partnerID := tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, partnerEmail, room.ID, testPool, authCtx.Redis, crypto)
	return reminderFixture{partnerID: partnerID, roomID: room.ID, productID: products[0].ID}
}

// book creates an awaiting-payment booking starting in `in`, for a guest
// (the guest email is unique per booking) in the given locale.
func (f reminderFixture) book(t *testing.T, ctx context.Context, in time.Duration, guestEmail string, locale domain.BookingLocale) *domain.Booking {
	t.Helper()
	start := time.Now().Add(in).Truncate(time.Minute)
	availability := tavail.NewTestAvailability(t, f.partnerID, f.roomID, start, start.Add(2*time.Hour), 0)
	availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
	require.NoError(t, err)
	tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

	b, err := service.CreateBooking(ctx, availability.ID, nil, f.productID, start,
		"Jane", "Doe", guestEmail, "+33612345678", locale)
	require.NoError(t, err)
	require.Equal(t, domain.BookingStatusAwaitingPayment, b.Status)
	return b
}

// waitForReminders waits until n reminders are recorded, then gives a
// duplicate the chance to show up before returning them all.
func waitForReminders(t *testing.T, n int) []notificationDomain.BookingReminderRequest {
	t.Helper()
	if n > 0 {
		require.Eventually(t, func() bool {
			return len(notificationEmailSpy.Reminders()) >= n
		}, 5*time.Second, 20*time.Millisecond, "expected %d reminder email(s)", n)
	}
	time.Sleep(300 * time.Millisecond)
	return notificationEmailSpy.Reminders()
}

func newEmailReminderScheduler() *bookingService.ReminderScheduler {
	return bookingService.NewReminderScheduler(bookingRepo, realNotificationAdapter, crypto,
		bookingService.WithReminderWindow(24*time.Hour))
}

// A confirmed booking 24 hours ahead gets exactly one reminder email, in its
// locale, through the real notification adapter, even when the scheduler
// runs again and when two instances tick at the same time; cancelled,
// expired and awaiting-payment bookings get none; no SMS (issue 019).
func TestReminderEmail_ConfirmedBookingsOnlyOnceInTheirLocale(t *testing.T) {
	ctx := context.Background()
	defer tu.ClearAuthData(t, ctx, authCtx)
	f := setupReminderFixture(t, ctx, "partner-reminder-email@example.com")

	french := f.book(t, ctx, 23*time.Hour+50*time.Minute, "fr.client@example.com", domain.BookingLocaleFR)
	confirmBookingPayment(t, ctx, french.ID)

	english := f.book(t, ctx, 20*time.Hour, "en.client@example.com", domain.BookingLocaleEN)
	confirmBookingPayment(t, ctx, english.ID)

	cancelled := f.book(t, ctx, 17*time.Hour, "cancelled@example.com", domain.BookingLocaleFR)
	confirmBookingPayment(t, ctx, cancelled.ID)
	require.NoError(t, tbooking.UpdateBookingStatus(t, ctx, testPool, cancelled.ID, domain.BookingStatusCancelled))

	expired := f.book(t, ctx, 14*time.Hour, "expired@example.com", domain.BookingLocaleFR)
	require.NoError(t, tbooking.UpdateBookingStatus(t, ctx, testPool, expired.ID, domain.BookingStatusExpired))

	awaiting := f.book(t, ctx, 11*time.Hour, "awaiting@example.com", domain.BookingLocaleFR)

	notificationEmailSpy.Reset()
	smsBefore := notificationSMSSpy.Count()

	// Two instances tick at the same time, then one ticks again.
	first, second := newEmailReminderScheduler(), newEmailReminderScheduler()
	var wg sync.WaitGroup
	for _, s := range []*bookingService.ReminderScheduler{first, second} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.TickOnce(ctx)
		}()
	}
	wg.Wait()
	first.TickOnce(ctx)

	reminders := waitForReminders(t, 2)
	require.Len(t, reminders, 2, "one reminder per confirmed booking, sent once")

	byBooking := map[string]notificationDomain.BookingReminderRequest{}
	for _, r := range reminders {
		byBooking[r.BookingID] = r
	}
	require.Contains(t, byBooking, french.ID.String())
	require.Contains(t, byBooking, english.ID.String())
	for _, id := range []uuid.UUID{cancelled.ID, expired.ID, awaiting.ID} {
		assert.NotContains(t, byBooking, id.String(), "only confirmed bookings are reminded")
	}

	fr := byBooking[french.ID.String()]
	assert.Equal(t, notificationDomain.LocaleFR, fr.Locale)
	assert.Equal(t, "fr.client@example.com", fr.ToEmail)

	en := byBooking[english.ID.String()]
	assert.Equal(t, notificationDomain.LocaleEN, en.Locale, "the reminder follows the booking's locale")
	assert.Equal(t, "en.client@example.com", en.ToEmail)
	assert.Equal(t, "Jane", en.ToFirstName)
	assert.Equal(t, stubPracticeAddress, en.Address, "the practice address comes from the company settings")
	assert.NotEmpty(t, en.ProductName)
	assert.True(t, en.StartTime.Equal(english.SlotStartTime))
	assert.True(t, en.EndTime.Equal(english.SlotEndTime))
	manage, err := url.Parse(en.ManageURL)
	require.NoError(t, err)
	assert.Equal(t, "https://app.leviosa.test", manage.Scheme+"://"+manage.Host)
	assert.Equal(t, "/bookings", manage.Path)
	assert.Equal(t, english.Token, manage.Query().Get("token"))

	// The English reminder is composed with the English copy.
	composed := notificationDomain.ComposeBookingReminder(en, time.Now())
	assert.Contains(t, composed.Subject, "Reminder: your appointment on")

	assert.Equal(t, smsBefore, notificationSMSSpy.Count(), "booking reminders send no SMS")

	for _, id := range []uuid.UUID{cancelled.ID, expired.ID, awaiting.ID} {
		encx, err := bookingRepo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Nil(t, encx.RemindedAt, "a booking that got no reminder is not marked reminded")
	}
}

// cancelAfterQueryRepo cancels a booking right after the scheduler's query
// returned it, as a client cancelling during a tick would.
type cancelAfterQueryRepo struct {
	ports.BookingRepository
	t      *testing.T
	cancel uuid.UUID
}

func (r *cancelAfterQueryRepo) FindBookingsDueForReminder(ctx context.Context) ([]*domain.BookingEncx, error) {
	bookings, err := r.BookingRepository.FindBookingsDueForReminder(ctx)
	require.NoError(r.t, tbooking.UpdateBookingStatus(r.t, ctx, testPool, r.cancel, domain.BookingStatusCancelled))
	return bookings, err
}

// A booking cancelled after the scheduler picked it up gets no reminder.
func TestReminderEmail_CancelledDuringTickGetsNone(t *testing.T) {
	ctx := context.Background()
	defer tu.ClearAuthData(t, ctx, authCtx)
	f := setupReminderFixture(t, ctx, "partner-reminder-cancel@example.com")

	b := f.book(t, ctx, 23*time.Hour, "cancel.during@example.com", domain.BookingLocaleFR)
	confirmBookingPayment(t, ctx, b.ID)

	notificationEmailSpy.Reset()
	repo := &cancelAfterQueryRepo{BookingRepository: bookingRepo, t: t, cancel: b.ID}
	scheduler := bookingService.NewReminderSchedulerForTest(repo, realNotificationAdapter, crypto,
		bookingService.WithReminderWindow(24*time.Hour))
	scheduler.TickOnce(ctx)

	assert.Empty(t, waitForReminders(t, 0), "a booking cancelled during the tick is not reminded")
}
