package booking

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
	"github.com/hengadev/encx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockPractitionerRepo struct {
	mu         sync.Mutex
	bookings   []*domain.BookingEncx
	markCalled []uuid.UUID
	// practitionerClaimed holds the bookings already claimed; unclaimable
	// those whose claim fails as in Postgres when the booking is no longer
	// confirmed (cancelled, awaiting payment, expired).
	claimed     map[uuid.UUID]bool
	unclaimable map[uuid.UUID]bool
	// hidden are excluded from the query result, as Postgres does once
	// practitioner_reminded_at is set.
	hidden map[uuid.UUID]bool
}

func (m *mockPractitionerRepo) FindBookingsDueForPractitionerReminder(_ context.Context) ([]*domain.BookingEncx, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*domain.BookingEncx
	for _, b := range m.bookings {
		if !m.hidden[b.ID] {
			out = append(out, b)
		}
	}
	return out, nil
}

func (m *mockPractitionerRepo) MarkPractitionerReminderSent(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markCalled = append(m.markCalled, id)
	if m.hidden == nil {
		m.hidden = map[uuid.UUID]bool{}
	}
	m.hidden[id] = true
	return nil
}

// ClaimPractitionerReminder mimics the conditional update: a booking is
// claimed once, never when unclaimable. Successful claims hide the booking
// from the query, as practitioner_reminded_at no longer being NULL does.
func (m *mockPractitionerRepo) ClaimPractitionerReminder(_ context.Context, id uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.claimed == nil {
		m.claimed = map[uuid.UUID]bool{}
	}
	if m.hidden == nil {
		m.hidden = map[uuid.UUID]bool{}
	}
	if m.claimed[id] || m.unclaimable[id] {
		return false, nil
	}
	m.claimed[id] = true
	m.hidden[id] = true
	return true, nil
}

type mockSessionReminderNotification struct {
	mu    sync.Mutex
	calls []ports.BookingNotificationData
	err   error
}

func (m *mockSessionReminderNotification) SendPractitionerSessionReminder(_ context.Context, data ports.BookingNotificationData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, data)
	return m.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makePractitionerEncryptedBooking(t *testing.T, crypto encx.CryptoService, slotStart time.Time, status domain.BookingStatus) *domain.BookingEncx {
	t.Helper()
	clientID := uuid.New()
	b := &domain.Booking{
		ID:              uuid.New(),
		AvailabilityID:  uuid.New(),
		ClientID:        &clientID,
		PartnerID:       uuid.New(),
		RoomID:          uuid.New(),
		ProductID:       uuid.New(),
		SlotStartTime:   slotStart,
		SlotEndTime:     slotStart.Add(time.Hour),
		TotalPriceCents: 5000,
		Currency:        "EUR",
		PaymentStatus:   domain.PaymentStatusPaid,
		Status:          status,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	encxBooking, err := domain.ProcessBookingEncx(context.Background(), crypto, b)
	require.NoError(t, err)
	return encxBooking
}

func newTestPractitionerScheduler(t *testing.T, repo *mockPractitionerRepo, notif *mockSessionReminderNotification) *PractitionerReminderScheduler {
	t.Helper()
	crypto := &identityCrypto{pepper: make([]byte, 32)}
	return NewPractitionerReminderSchedulerForTest(repo, notif, crypto, WithPractitionerReminderWindow(15*time.Minute))
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPractitionerReminderScheduler_SentOnceAcrossTwoTicks(t *testing.T) {
	ctx := context.Background()
	repo := &mockPractitionerRepo{}
	notif := &mockSessionReminderNotification{}
	s := newTestPractitionerScheduler(t, repo, notif)

	inWindow := makePractitionerEncryptedBooking(t, &identityCrypto{pepper: make([]byte, 32)}, time.Now().Add(10*time.Minute), domain.BookingStatusConfirmed)
	repo.bookings = []*domain.BookingEncx{inWindow}

	s.TickOnce(ctx)
	notif.mu.Lock()
	first := len(notif.calls)
	notif.mu.Unlock()
	assert.Equal(t, 1, first)

	// Second tick: the claimed booking no longer matches the query.
	s.TickOnce(ctx)
	notif.mu.Lock()
	assert.Len(t, notif.calls, 1)
	assert.Equal(t, inWindow.ID, notif.calls[0].BookingID)
	assert.Equal(t, *inWindow.ClientID, notif.calls[0].ClientID)
	notif.mu.Unlock()
}

func TestPractitionerReminderScheduler_NotResentAfterRestart(t *testing.T) {
	ctx := context.Background()
	crypto := &identityCrypto{pepper: make([]byte, 32)}
	repo := &mockPractitionerRepo{}
	notif := &mockSessionReminderNotification{}
	inWindow := makePractitionerEncryptedBooking(t, crypto, time.Now().Add(10*time.Minute), domain.BookingStatusConfirmed)
	repo.bookings = []*domain.BookingEncx{inWindow}

	newTestPractitionerScheduler(t, repo, notif).TickOnce(ctx)
	// A restart builds a fresh scheduler over the same persisted state: the
	// recorded time hides the booking from the query.
	newTestPractitionerScheduler(t, repo, notif).TickOnce(ctx)

	notif.mu.Lock()
	defer notif.mu.Unlock()
	assert.Len(t, notif.calls, 1)
}

func TestPractitionerReminderScheduler_NotSentForStartedBooking(t *testing.T) {
	ctx := context.Background()
	repo := &mockPractitionerRepo{}
	notif := &mockSessionReminderNotification{}
	started := makePractitionerEncryptedBooking(t, &identityCrypto{pepper: make([]byte, 32)}, time.Now().Add(-1*time.Minute), domain.BookingStatusConfirmed)
	repo.bookings = []*domain.BookingEncx{started}

	newTestPractitionerScheduler(t, repo, notif).TickOnce(ctx)

	notif.mu.Lock()
	defer notif.mu.Unlock()
	assert.Empty(t, notif.calls)
	// Marked without sending so it does not reappear.
	repo.mu.Lock()
	defer repo.mu.Unlock()
	assert.Contains(t, repo.markCalled, started.ID)
}

func TestPractitionerReminderScheduler_NotSentOutsideWindow(t *testing.T) {
	ctx := context.Background()
	repo := &mockPractitionerRepo{}
	notif := &mockSessionReminderNotification{}
	later := makePractitionerEncryptedBooking(t, &identityCrypto{pepper: make([]byte, 32)}, time.Now().Add(2*time.Hour), domain.BookingStatusConfirmed)
	repo.bookings = []*domain.BookingEncx{later}

	newTestPractitionerScheduler(t, repo, notif).TickOnce(ctx)

	notif.mu.Lock()
	defer notif.mu.Unlock()
	assert.Empty(t, notif.calls)
}

func TestPractitionerReminderScheduler_NotSentForNonConfirmedBookings(t *testing.T) {
	ctx := context.Background()
	for _, status := range []domain.BookingStatus{
		domain.BookingStatusCancelled,
		domain.BookingStatusExpired,
		domain.BookingStatusAwaitingPayment,
	} {
		repo := &mockPractitionerRepo{}
		notif := &mockSessionReminderNotification{}
		b := makePractitionerEncryptedBooking(t, &identityCrypto{pepper: make([]byte, 32)}, time.Now().Add(10*time.Minute), status)
		// Non-confirmed bookings never match the query; if one slipped
		// through anyway, its claim fails like Postgres' conditional update.
		repo.bookings = []*domain.BookingEncx{b}
		repo.unclaimable = map[uuid.UUID]bool{b.ID: true}

		newTestPractitionerScheduler(t, repo, notif).TickOnce(ctx)

		notif.mu.Lock()
		calls := len(notif.calls)
		notif.mu.Unlock()
		assert.Empty(t, calls, "status %s", status)
	}
}

func TestPractitionerReminderScheduler_SentForGuestBookingWithNames(t *testing.T) {
	ctx := context.Background()
	crypto := &identityCrypto{pepper: make([]byte, 32)}
	repo := &mockPractitionerRepo{}
	notif := &mockSessionReminderNotification{}
	b, err := domain.NewBooking(
		uuid.New(), nil, uuid.New(), uuid.New(),
		5000, "EUR",
		"Jean", "Dupont", "jean@example.com", "+33612345678",
	)
	require.NoError(t, err)
	b.SlotStartTime = time.Now().Add(10 * time.Minute)
	b.SlotEndTime = b.SlotStartTime.Add(time.Hour)
	b.ProductID = uuid.New()
	b.Status = domain.BookingStatusConfirmed
	b.PaymentStatus = domain.PaymentStatusPaid
	encxBooking, err := domain.ProcessBookingEncx(ctx, crypto, b)
	require.NoError(t, err)
	repo.bookings = []*domain.BookingEncx{encxBooking}

	newTestPractitionerScheduler(t, repo, notif).TickOnce(ctx)

	notif.mu.Lock()
	defer notif.mu.Unlock()
	require.Len(t, notif.calls, 1)
	assert.True(t, notif.calls[0].IsGuestBooking)
	assert.Equal(t, "Jean", notif.calls[0].ClientFirstName)
	assert.Equal(t, "Dupont", notif.calls[0].ClientLastName)
}
