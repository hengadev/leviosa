package booking

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// accessUpdateRepo returns its bookings from GetUpcoming; any other
// repository method panics (nil embedded interface).
type accessUpdateRepo struct {
	ports.BookingRepository
	bookings []*domain.BookingEncx
}

func (r *accessUpdateRepo) GetUpcoming(_ context.Context, _ ports.BookingFilter) ([]*domain.BookingEncx, error) {
	return r.bookings, nil
}

// accessUpdateUsers resolves registered clients' emails from a map; a
// client missing from it has no resolvable email.
type accessUpdateUsers struct {
	ports.AuthUserClient
	emails map[uuid.UUID]string
}

func (u *accessUpdateUsers) GetUserEmail(_ context.Context, id uuid.UUID) (string, error) {
	email, ok := u.emails[id]
	if !ok {
		return "", errors.New("user not found")
	}
	return email, nil
}

type fixedPracticeAccess struct {
	access domain.PracticeAccess
	err    error
}

func (p fixedPracticeAccess) GetPracticeAccess(context.Context, uuid.UUID) (domain.PracticeAccess, error) {
	return p.access, p.err
}

type recordingAccessNotifier struct {
	mu    sync.Mutex
	calls []ports.BookingNotificationData
}

func (n *recordingAccessNotifier) SendBookingAccessUpdate(_ context.Context, data ports.BookingNotificationData) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, data)
	return nil
}

type recordingAuditLogger struct {
	entries []audit.Entry
}

func (l *recordingAuditLogger) Write(_ context.Context, e audit.Entry) error {
	l.entries = append(l.entries, e)
	return nil
}

var accessUpdateNow = time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)

type accessUpdateFixture struct {
	svc      *BookingService
	repo     *accessUpdateRepo
	users    *accessUpdateUsers
	notifier *recordingAccessNotifier
	audit    *recordingAuditLogger
}

func newAccessUpdateFixture(doorCode string) *accessUpdateFixture {
	f := &accessUpdateFixture{
		repo:     &accessUpdateRepo{},
		users:    &accessUpdateUsers{emails: map[uuid.UUID]string{}},
		notifier: &recordingAccessNotifier{},
		audit:    &recordingAuditLogger{},
	}
	f.svc = &BookingService{
		bookingRepo:    f.repo,
		crypto:         &identityCrypto{pepper: make([]byte, 32)},
		authUserClient: f.users,
		practiceAccess: fixedPracticeAccess{access: domain.PracticeAccess{DoorCode: doorCode, Instructions: "2e étage"}},
		accessNotifier: f.notifier,
		audit:          f.audit,
		clock:          func() time.Time { return accessUpdateNow },
	}
	return f
}

// addGuest adds a guest booking with the given email (may be empty).
func (f *accessUpdateFixture) addGuest(t *testing.T, email string, start time.Time, status domain.BookingStatus) uuid.UUID {
	t.Helper()
	b := &domain.Booking{
		ID:             uuid.New(),
		AvailabilityID: uuid.New(),
		PartnerID:      uuid.New(),
		RoomID:         uuid.New(),
		ProductID:      uuid.New(),
		GuestFirstName: "Jean",
		GuestLastName:  "Dupont",
		GuestEmail:     email,
		SlotStartTime:  start,
		SlotEndTime:    start.Add(time.Hour),
		Currency:       "EUR",
		Status:         status,
		Locale:         domain.BookingLocaleEN,
	}
	f.add(t, b)
	return b.ID
}

// addClient adds a registered client's booking; email "" leaves the client
// without a resolvable email.
func (f *accessUpdateFixture) addClient(t *testing.T, email string, start time.Time) uuid.UUID {
	t.Helper()
	clientID := uuid.New()
	if email != "" {
		f.users.emails[clientID] = email
	}
	b := &domain.Booking{
		ID:             uuid.New(),
		AvailabilityID: uuid.New(),
		ClientID:       &clientID,
		PartnerID:      uuid.New(),
		RoomID:         uuid.New(),
		ProductID:      uuid.New(),
		SlotStartTime:  start,
		SlotEndTime:    start.Add(time.Hour),
		Currency:       "EUR",
		Status:         domain.BookingStatusConfirmed,
	}
	f.add(t, b)
	return b.ID
}

func (f *accessUpdateFixture) add(t *testing.T, b *domain.Booking) {
	t.Helper()
	bx, err := domain.ProcessBookingEncx(context.Background(), f.svc.crypto, b)
	require.NoError(t, err)
	f.repo.bookings = append(f.repo.bookings, bx)
}

func TestGetAccessUpdateRecipients_CountsUpcomingConfirmedBookings(t *testing.T) {
	f := newAccessUpdateFixture("4821B")
	tomorrow := accessUpdateNow.Add(24 * time.Hour)
	f.addGuest(t, "guest@example.com", tomorrow, domain.BookingStatusConfirmed)
	f.addClient(t, "client@example.com", tomorrow.Add(time.Hour))
	f.addGuest(t, "", tomorrow, domain.BookingStatusConfirmed)                                                // guest without email
	f.addClient(t, "", tomorrow)                                                                              // email not resolvable
	f.addGuest(t, "started@example.com", accessUpdateNow.Add(-30*time.Minute), domain.BookingStatusConfirmed) // already started
	f.addGuest(t, "past@example.com", accessUpdateNow.Add(-48*time.Hour), domain.BookingStatusConfirmed)
	f.addGuest(t, "cancelled@example.com", tomorrow, domain.BookingStatusCancelled)

	counts, err := f.svc.GetAccessUpdateRecipients(context.Background())
	require.NoError(t, err)
	assert.Equal(t, &domain.AccessUpdateRecipients{Recipients: 2, WithoutEmail: 2}, counts)
	assert.Empty(t, f.notifier.calls, "counting sends nothing")
	assert.Empty(t, f.audit.entries, "counting is not audited")
}

func TestSendAccessUpdate_SendsToUpcomingBookingsWithEmail(t *testing.T) {
	f := newAccessUpdateFixture("4821B")
	tomorrow := accessUpdateNow.Add(24 * time.Hour)
	guest := f.addGuest(t, "guest@example.com", tomorrow, domain.BookingStatusConfirmed)
	client := f.addClient(t, "client@example.com", tomorrow.Add(time.Hour))
	f.addGuest(t, "", tomorrow, domain.BookingStatusConfirmed)
	f.addClient(t, "", tomorrow)
	f.addGuest(t, "started@example.com", accessUpdateNow.Add(-time.Minute), domain.BookingStatusConfirmed)
	f.addGuest(t, "noshow@example.com", tomorrow, domain.BookingStatusNoShow)

	result, err := f.svc.SendAccessUpdate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, &domain.AccessUpdateRecipients{Recipients: 2, WithoutEmail: 2}, result)

	sent := map[uuid.UUID]ports.BookingNotificationData{}
	for _, d := range f.notifier.calls {
		sent[d.BookingID] = d
	}
	require.Len(t, sent, 2)
	assert.Equal(t, "guest@example.com", sent[guest].ClientEmail)
	assert.Equal(t, "en", sent[guest].Locale)
	assert.Equal(t, "client@example.com", sent[client].ClientEmail)
	assert.False(t, sent[client].IsGuestBooking)

	require.Len(t, f.audit.entries, 1)
	entry := f.audit.entries[0]
	assert.Equal(t, "booking.access_update_sent", entry.Action)
	assert.Equal(t, map[string]any{"recipients": 2, "without_email": 2}, entry.Payload)
	assert.NotContains(t, fmt.Sprint(entry), "4821B", "the audit entry never carries the code")
}

func TestSendAccessUpdate_NoBookingsSendsNothing(t *testing.T) {
	f := newAccessUpdateFixture("4821B")

	result, err := f.svc.SendAccessUpdate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, &domain.AccessUpdateRecipients{}, result)
	assert.Empty(t, f.notifier.calls)
}

func TestSendAccessUpdate_RefusedWithoutDoorCode(t *testing.T) {
	for name, code := range map[string]string{"empty": "", "blank": "   "} {
		t.Run(name, func(t *testing.T) {
			f := newAccessUpdateFixture(code)
			f.addGuest(t, "guest@example.com", accessUpdateNow.Add(24*time.Hour), domain.BookingStatusConfirmed)

			_, err := f.svc.SendAccessUpdate(context.Background())
			require.Error(t, err)
			assert.True(t, errors.Is(err, errs.ErrInvalidValue), "got %v", err)
			assert.Empty(t, f.notifier.calls)
			assert.Empty(t, f.audit.entries)
		})
	}
}

func TestSendAccessUpdate_RefusedWithoutPracticeAccessProvider(t *testing.T) {
	f := newAccessUpdateFixture("4821B")
	f.svc.practiceAccess = nil
	f.addGuest(t, "guest@example.com", accessUpdateNow.Add(24*time.Hour), domain.BookingStatusConfirmed)

	_, err := f.svc.SendAccessUpdate(context.Background())
	assert.True(t, errors.Is(err, errs.ErrInvalidValue), "got %v", err)
	assert.Empty(t, f.notifier.calls)
}

func TestSendAccessUpdate_PracticeAccessReadErrorIsNotInvalidValue(t *testing.T) {
	f := newAccessUpdateFixture("")
	f.svc.practiceAccess = fixedPracticeAccess{err: errors.New("settings down")}

	_, err := f.svc.SendAccessUpdate(context.Background())
	require.Error(t, err)
	assert.False(t, errors.Is(err, errs.ErrInvalidValue))
	assert.False(t, strings.Contains(err.Error(), "4821B"))
}
