package booking

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	catalogPorts "github.com/Leviosa-care/leviosa/backend/internal/catalog/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Practitioner booking rules (issue 041, ADR-0021), with fakes: the slot
// engine's rules without the minimum notice and horizon, confirmed with
// nothing paid online, no client message, cancelled without a refund.

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

// practitionerBookingRepo keeps bookings in memory. CreateGuarded serialises
// like the partner's agenda lock; any method not overridden panics.
type practitionerBookingRepo struct {
	ports.BookingRepository
	mu       sync.Mutex
	bookings []*domain.BookingEncx
}

func (r *practitionerBookingRepo) occupying(partnerID uuid.UUID) []*domain.BookingEncx {
	var out []*domain.BookingEncx
	for _, b := range r.bookings {
		if b.PartnerID == partnerID && b.Status != domain.BookingStatusCancelled && b.Status != domain.BookingStatusExpired {
			out = append(out, b)
		}
	}
	return out
}

func (r *practitionerBookingRepo) GetByPartnerID(_ context.Context, partnerID uuid.UUID, _ ports.BookingFilter) ([]*domain.BookingEncx, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.occupying(partnerID), nil
}

func (r *practitionerBookingRepo) CreateGuarded(_ context.Context, b *domain.BookingEncx, verify func([]*domain.BookingEncx) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := verify(r.occupying(b.PartnerID)); err != nil {
		return err
	}
	r.bookings = append(r.bookings, b)
	return nil
}

func (r *practitionerBookingRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.BookingEncx, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range r.bookings {
		if b.ID == id {
			return b, nil
		}
	}
	return nil, errs.ErrRepositoryNotFound
}

func (r *practitionerBookingRepo) ClaimCancellation(_ context.Context, id uuid.UUID, by domain.CancellationActor, at time.Time, refund int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range r.bookings {
		if b.ID == id && b.Status == domain.BookingStatusConfirmed {
			b.Status = domain.BookingStatusCancelled
			b.CancelledAt = &at
			actor := string(by)
			b.CancelledBy = &actor
			b.RefundedAmountCents = &refund
			return true, nil
		}
	}
	return false, nil
}

func (r *practitionerBookingRepo) GetUpcoming(_ context.Context, _ ports.BookingFilter) ([]*domain.BookingEncx, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*domain.BookingEncx(nil), r.bookings...), nil
}

type practitionerAvailabilityRepo struct {
	ports.AvailabilityRepository
	availabilities map[uuid.UUID]*domain.AvailabilityEncx
}

func (r *practitionerAvailabilityRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.AvailabilityEncx, error) {
	if a, ok := r.availabilities[id]; ok {
		return a, nil
	}
	return nil, errs.ErrRepositoryNotFound
}

func (r *practitionerAvailabilityRepo) GetByUserID(_ context.Context, userID uuid.UUID, _ ports.AvailabilityFilter) ([]*domain.AvailabilityEncx, error) {
	var out []*domain.AvailabilityEncx
	for _, a := range r.availabilities {
		if a.UserID == userID {
			out = append(out, a)
		}
	}
	return out, nil
}

// Releasing the availability on cancellation is best effort.
func (r *practitionerAvailabilityRepo) Update(context.Context, *domain.AvailabilityEncx) error {
	return nil
}

type practitionerProducts struct {
	products map[uuid.UUID]*catalogDomain.ProductRes
}

func (p *practitionerProducts) GetProductByID(_ context.Context, id string) (*catalogDomain.ProductRes, error) {
	for _, product := range p.products {
		if product.ID.String() == id {
			return product, nil
		}
	}
	return nil, errs.ErrRepositoryNotFound
}

func (p *practitionerProducts) GetAllPublishedProducts(context.Context) ([]*catalogDomain.ProductRes, error) {
	var out []*catalogDomain.ProductRes
	for _, product := range p.products {
		if product.Status == catalogDomain.Published {
			out = append(out, product)
		}
	}
	return out, nil
}

func (p *practitionerProducts) GetAllProducts(context.Context) ([]*catalogDomain.ProductRes, error) {
	return nil, nil
}

type practitionerPrices struct {
	catalogPorts.PublicPriceService
}

func (practitionerPrices) GetActiveOneTimePriceByProductID(context.Context, string, string) (*catalogDomain.Price, error) {
	return &catalogDomain.Price{Amount: 7000, Currency: "EUR", IsActive: true}, nil
}

// recordingPayments records every call: a practitioner booking makes none.
type recordingPayments struct {
	ports.PaymentService
	mu    sync.Mutex
	calls []string
}

func (p *recordingPayments) record(call string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, call)
}

func (p *recordingPayments) CreatePaymentIntent(context.Context, int, string, string, string, map[string]string) (string, string, error) {
	p.record("create")
	return "pi_test", "secret", nil
}

func (p *recordingPayments) RefundPayment(context.Context, string, int, string) (string, error) {
	p.record("refund")
	return "re_test", nil
}

func (p *recordingPayments) CancelPaymentIntent(context.Context, string) error {
	p.record("cancel")
	return nil
}

// recordingNotifications records every booking message by name.
type recordingNotifications struct {
	mu    sync.Mutex
	calls []string
}

func (n *recordingNotifications) record(name string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, name)
	return nil
}

func (n *recordingNotifications) SendBookingConfirmation(context.Context, ports.BookingNotificationData) error {
	return n.record("confirmation")
}
func (n *recordingNotifications) SendBookingCancellation(context.Context, ports.BookingNotificationData) error {
	return n.record("cancellation")
}
func (n *recordingNotifications) SendBookingReminder(context.Context, ports.BookingNotificationData) error {
	return n.record("reminder")
}
func (n *recordingNotifications) SendPaymentConfirmation(context.Context, ports.BookingNotificationData) error {
	return n.record("payment confirmation")
}
func (n *recordingNotifications) SendPaymentFailed(context.Context, ports.BookingNotificationData) error {
	return n.record("payment failed")
}
func (n *recordingNotifications) SendPaymentDisputeAlert(context.Context, ports.BookingNotificationData) error {
	return n.record("dispute")
}
func (n *recordingNotifications) SendBookingAccessUpdate(context.Context, ports.BookingNotificationData) error {
	return n.record("access update")
}

func (n *recordingNotifications) snapshot() []string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return append([]string(nil), n.calls...)
}

// practitionerUsers resolves the practitioner's partner profile (by partner
// ID and by user ID) with the products it offers.
type practitionerUsers struct {
	ports.AuthUserClient
	partnerID  uuid.UUID
	userID     uuid.UUID
	productIDs []uuid.UUID
}

func (u *practitionerUsers) GetPartnerUserID(_ context.Context, partnerID uuid.UUID) (uuid.UUID, error) {
	if partnerID != u.partnerID {
		return uuid.Nil, errs.ErrRepositoryNotFound
	}
	return u.userID, nil
}

func (u *practitionerUsers) GetPartnerByUserID(_ context.Context, userID uuid.UUID) (*ports.PartnerInfo, error) {
	if userID != u.userID {
		return nil, errs.ErrRepositoryNotFound
	}
	return &ports.PartnerInfo{ID: u.partnerID, UserID: u.userID, IsVerified: true, ProductIDs: u.productIDs}, nil
}

func (u *practitionerUsers) GetUserEmail(context.Context, uuid.UUID) (string, error) {
	return "", errs.ErrRepositoryNotFound
}

// ---------------------------------------------------------------------------
// Fixture
// ---------------------------------------------------------------------------

type practitionerFixture struct {
	svc           *BookingService
	repo          *practitionerBookingRepo
	availability  *domain.Availability
	product       *catalogDomain.ProductRes
	payments      *recordingPayments
	notifications *recordingNotifications
	users         *practitionerUsers
	now           time.Time
	// start is the availability's start: 09:00 two hours from now, well
	// within the default 12-hour minimum notice.
	start time.Time
}

func newPractitionerFixture(t *testing.T) *practitionerFixture {
	t.Helper()
	crypto := &identityCrypto{pepper: make([]byte, 32)}
	now := time.Now().Truncate(time.Minute)
	start := now.Add(2 * time.Hour).Truncate(time.Hour)
	userID := uuid.New()

	availability := &domain.Availability{
		ID:          uuid.New(),
		UserID:      userID,
		RoomID:      uuid.New(),
		StartTime:   start,
		EndTime:     start.Add(4 * time.Hour),
		MaxCapacity: 1,
		Status:      domain.AvailabilityStatusAvailable,
	}
	availabilityEncx, err := domain.ProcessAvailabilityEncx(context.Background(), crypto, availability)
	require.NoError(t, err)

	product := &catalogDomain.ProductRes{ID: uuid.New(), Name: "Séance d'ostéopathie", Duration: 60, Status: catalogDomain.Published}
	draft := &catalogDomain.ProductRes{ID: uuid.New(), Name: "Brouillon", Duration: 60, Status: catalogDomain.Draft}

	f := &practitionerFixture{
		repo:          &practitionerBookingRepo{},
		availability:  availability,
		product:       product,
		payments:      &recordingPayments{},
		notifications: &recordingNotifications{},
		users:         &practitionerUsers{partnerID: uuid.New(), userID: userID, productIDs: []uuid.UUID{product.ID, draft.ID}},
		now:           now,
		start:         start,
	}
	f.svc = &BookingService{
		bookingRepo:         f.repo,
		availabilityRepo:    &practitionerAvailabilityRepo{availabilities: map[uuid.UUID]*domain.AvailabilityEncx{availability.ID: availabilityEncx}},
		paymentService:      f.payments,
		productService:      &practitionerProducts{products: map[uuid.UUID]*catalogDomain.ProductRes{product.ID: product, draft.ID: draft}},
		priceService:        practitionerPrices{},
		notificationService: f.notifications,
		accessNotifier:      f.notifications,
		practiceAccess:      fixedPracticeAccess{access: domain.PracticeAccess{DoorCode: "4821B"}},
		authUserClient:      f.users,
		crypto:              crypto,
		tokenSecret:         []byte("test-token-secret-32-bytes-long!"),
		clock:               func() time.Time { return now },
	}
	return f
}

func (f *practitionerFixture) input(start time.Time) domain.PractitionerBookingInput {
	return domain.PractitionerBookingInput{
		AvailabilityID: f.availability.ID,
		ProductID:      f.product.ID,
		SlotStartTime:  start,
		GuestFirstName: "Marie",
		GuestLastName:  "Dupont",
		GuestPhone:     "0612345678",
	}
}

func (f *practitionerFixture) book(t *testing.T, start time.Time) *domain.Booking {
	t.Helper()
	booking, err := f.svc.CreatePractitionerBooking(context.Background(), f.input(start), f.availability.UserID, false)
	require.NoError(t, err)
	return booking
}

// ---------------------------------------------------------------------------
// Slot rules
// ---------------------------------------------------------------------------

func TestCreatePractitionerBooking_SlotWithinMinimumNoticeIsAccepted(t *testing.T) {
	f := newPractitionerFixture(t)
	require.Less(t, f.start.Sub(f.now), domain.DefaultMinimumNotice, "the slot must be within the minimum notice")

	booking := f.book(t, f.start)

	assert.Equal(t, domain.BookingStatusConfirmed, booking.Status)
	assert.Equal(t, domain.BookingOriginPractitioner, booking.Origin)
	require.NotNil(t, booking.AmountPaidOnlineCents)
	assert.Equal(t, 0, *booking.AmountPaidOnlineCents)
	assert.Equal(t, 7000, booking.TotalPriceCents)
	assert.Equal(t, 7000, booking.AmountDueOnSiteCents(), "the full price is due at the practice")
	assert.Equal(t, f.start.Add(time.Hour), booking.SlotEndTime)
	assert.Nil(t, booking.PaymentIntentID, "no PaymentIntent")
	assert.Nil(t, booking.HoldExpiresAt, "no payment hold")
	assert.Empty(t, booking.Token, "no booking token is put to use")
	assert.Equal(t, "Marie", booking.GuestFirstName)

	assert.Empty(t, f.payments.calls, "nothing goes through the payment provider")
	assert.Empty(t, f.notifications.snapshot(), "no confirmation is sent")

	require.Len(t, f.repo.bookings, 1)
	stored, err := domain.DecryptBookingEncx(context.Background(), f.svc.crypto, f.repo.bookings[0])
	require.NoError(t, err)
	assert.Equal(t, domain.BookingOriginPractitioner, stored.Origin)
	assert.Equal(t, "0612345678", stored.GuestPhone)
}

func TestCreatePractitionerBooking_RefusesSlotsTheEngineDoesNotOffer(t *testing.T) {
	cases := map[string]func(f *practitionerFixture) time.Time{
		"before the availability": func(f *practitionerFixture) time.Time { return f.start.Add(-time.Hour) },
		"after the availability":  func(f *practitionerFixture) time.Time { return f.start.Add(4 * time.Hour) },
		"not fitting the end":     func(f *practitionerFixture) time.Time { return f.start.Add(3*time.Hour + 30*time.Minute) },
		"off the slot grid":       func(f *practitionerFixture) time.Time { return f.start.Add(30 * time.Minute) },
	}
	for name, start := range cases {
		t.Run(name, func(t *testing.T) {
			f := newPractitionerFixture(t)
			_, err := f.svc.CreatePractitionerBooking(context.Background(), f.input(start(f)), f.availability.UserID, false)
			require.ErrorIs(t, err, errs.ErrInvalidValue)
			assert.Empty(t, f.repo.bookings)
		})
	}
}

func TestCreatePractitionerBooking_RefusesAnOverlappingSlot(t *testing.T) {
	f := newPractitionerFixture(t)
	f.book(t, f.start.Add(time.Hour))

	_, err := f.svc.CreatePractitionerBooking(context.Background(), f.input(f.start.Add(time.Hour)), f.availability.UserID, false)
	require.ErrorIs(t, err, errs.ErrConflict)
	assert.Len(t, f.repo.bookings, 1)
}

func TestCreatePractitionerBooking_RefusesASlotThatHasStarted(t *testing.T) {
	f := newPractitionerFixture(t)
	f.svc.clock = func() time.Time { return f.start.Add(10 * time.Minute) }

	_, err := f.svc.CreatePractitionerBooking(context.Background(), f.input(f.start), f.availability.UserID, false)
	require.ErrorIs(t, err, errs.ErrInvalidValue)

	// A later slot of the occurrence that has started is still bookable.
	booking, err := f.svc.CreatePractitionerBooking(context.Background(), f.input(f.start.Add(time.Hour)), f.availability.UserID, false)
	require.NoError(t, err)
	assert.Equal(t, domain.BookingStatusConfirmed, booking.Status)
}

func TestCreatePractitionerBooking_WhoMayBook(t *testing.T) {
	f := newPractitionerFixture(t)

	_, err := f.svc.CreatePractitionerBooking(context.Background(), f.input(f.start), uuid.New(), false)
	require.ErrorIs(t, err, errs.ErrForbidden, "another partner")

	booking, err := f.svc.CreatePractitionerBooking(context.Background(), f.input(f.start), uuid.New(), true)
	require.NoError(t, err, "an administrator books for the practitioner")
	assert.Equal(t, f.availability.UserID, booking.PartnerID)
}

func TestCreatePractitionerBooking_Client(t *testing.T) {
	f := newPractitionerFixture(t)

	missingPhone := f.input(f.start)
	missingPhone.GuestPhone = "  "
	_, err := f.svc.CreatePractitionerBooking(context.Background(), missingPhone, f.availability.UserID, false)
	require.ErrorIs(t, err, errs.ErrInvalidValue)

	clientID := uuid.New()
	both := f.input(f.start)
	both.ClientID = &clientID
	_, err = f.svc.CreatePractitionerBooking(context.Background(), both, f.availability.UserID, false)
	require.ErrorIs(t, err, errs.ErrInvalidValue, "an existing client and guest fields are exclusive")

	existing := domain.PractitionerBookingInput{AvailabilityID: f.availability.ID, ProductID: f.product.ID, SlotStartTime: f.start, ClientID: &clientID}
	booking, err := f.svc.CreatePractitionerBooking(context.Background(), existing, f.availability.UserID, false)
	require.NoError(t, err)
	assert.Equal(t, &clientID, booking.ClientID)

	withEmail := f.input(f.start.Add(time.Hour))
	withEmail.GuestEmail = " marie@example.com "
	booking, err = f.svc.CreatePractitionerBooking(context.Background(), withEmail, f.availability.UserID, false)
	require.NoError(t, err)
	assert.Equal(t, "marie@example.com", booking.GuestEmail)
	assert.Empty(t, f.notifications.snapshot(), "an email does not make the app write to the client")
}

func TestCreatePractitionerBooking_OnlyPublishedProductsOfThePractitioner(t *testing.T) {
	f := newPractitionerFixture(t)

	var draftID uuid.UUID
	for _, id := range f.users.productIDs {
		if id != f.product.ID {
			draftID = id
		}
	}
	draft := f.input(f.start)
	draft.ProductID = draftID
	_, err := f.svc.CreatePractitionerBooking(context.Background(), draft, f.availability.UserID, false)
	require.ErrorIs(t, err, errs.ErrConflict, "a draft product")

	f.users.productIDs = []uuid.UUID{draftID}
	_, err = f.svc.CreatePractitionerBooking(context.Background(), f.input(f.start), f.availability.UserID, false)
	require.ErrorIs(t, err, errs.ErrConflict, "a product the practitioner does not offer")

	products, err := f.svc.GetPractitionerProducts(context.Background(), f.availability.UserID)
	require.NoError(t, err)
	assert.Empty(t, products)

	f.users.productIDs = []uuid.UUID{f.product.ID, draftID}
	products, err = f.svc.GetPractitionerProducts(context.Background(), f.availability.UserID)
	require.NoError(t, err)
	assert.Equal(t, []domain.PractitionerProduct{{ID: f.product.ID, Name: f.product.Name, DurationMinutes: 60, PriceCents: 7000, Currency: "EUR"}}, products)
}

// ---------------------------------------------------------------------------
// Free slots: the form's, and the public page's unchanged
// ---------------------------------------------------------------------------

func slotStarts(slots *domain.FreeSlots) []time.Time {
	var starts []time.Time
	for _, slot := range slots.Slots {
		starts = append(starts, slot.Start)
	}
	return starts
}

func TestFreeSlots_PractitionerWithoutNoticePublicUnchanged(t *testing.T) {
	f := newPractitionerFixture(t)
	f.book(t, f.start.Add(time.Hour))
	from, to := f.now, f.now.Add(48*time.Hour)

	practitioner, err := f.svc.GetPractitionerFreeSlots(context.Background(), f.availability.UserID, f.product.ID, from, to)
	require.NoError(t, err)
	assert.Equal(t, []time.Time{f.start, f.start.Add(2 * time.Hour), f.start.Add(3 * time.Hour)}, slotStarts(practitioner),
		"every free slot of the occurrence, the booked one excepted, despite the minimum notice")

	public, err := f.svc.GetFreeSlots(context.Background(), f.users.partnerID, f.product.ID, from, to)
	require.NoError(t, err)
	assert.Empty(t, public.Slots, "the practice page still applies the 12-hour minimum notice")
	assert.False(t, public.EarliestStart.Before(f.now.Add(domain.DefaultMinimumNotice)),
		"the public answer keeps the policy window")
}

// ---------------------------------------------------------------------------
// No client message, and cancellation without refund
// ---------------------------------------------------------------------------

func TestCancelPractitionerBooking_RefundsNothingAndEmailsNoOne(t *testing.T) {
	f := newPractitionerFixture(t)
	booking := f.book(t, f.start)

	cancelled, err := f.svc.CancelBookingByPractitioner(context.Background(), booking.ID, f.availability.UserID, false)
	require.NoError(t, err)

	assert.Equal(t, domain.BookingStatusCancelled, cancelled.Status)
	require.NotNil(t, cancelled.RefundedAmountCents)
	assert.Equal(t, 0, *cancelled.RefundedAmountCents)
	assert.Empty(t, f.payments.calls, "nothing to refund: the payment provider is not called")
	assert.Empty(t, f.notifications.snapshot(), "no cancellation email")

	// The slot is free again.
	slots, err := f.svc.GetPractitionerFreeSlots(context.Background(), f.availability.UserID, f.product.ID, f.now, f.now.Add(48*time.Hour))
	require.NoError(t, err)
	assert.True(t, slices.ContainsFunc(slots.Slots, func(s domain.Slot) bool { return s.Start.Equal(f.start) }))
}

func TestAccessUpdate_SkipsPractitionerBookings(t *testing.T) {
	f := newPractitionerFixture(t)
	withEmail := f.input(f.start)
	withEmail.GuestEmail = "marie@example.com"
	_, err := f.svc.CreatePractitionerBooking(context.Background(), withEmail, f.availability.UserID, false)
	require.NoError(t, err)

	counts, err := f.svc.GetAccessUpdateRecipients(context.Background())
	require.NoError(t, err)
	assert.Equal(t, &domain.AccessUpdateRecipients{}, counts, "neither a recipient nor a client without email")

	sent, err := f.svc.SendAccessUpdate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, sent.Recipients)
	assert.Empty(t, f.notifications.snapshot(), "the door code is never sent for it")
}

func TestReminders_PractitionerBooking(t *testing.T) {
	f := newPractitionerFixture(t)
	withEmail := f.input(f.start)
	withEmail.GuestEmail = "marie@example.com"
	_, err := f.svc.CreatePractitionerBooking(context.Background(), withEmail, f.availability.UserID, false)
	require.NoError(t, err)

	// The client's 24-hour reminder: none.
	repo := &mockSchedulerRepo{bookings: f.repo.bookings}
	clientReminders := &mockReminderNotification{}
	NewReminderSchedulerForTest(repo, clientReminders, f.svc.crypto).TickOnce(context.Background())
	assert.Empty(t, clientReminders.calls, "no reminder to the client")
	assert.Empty(t, repo.markCalled)

	// The practitioner's own 15-minute notification still fires.
	practitionerRepo := &mockPractitionerRepo{bookings: f.repo.bookings}
	sessionReminders := &mockSessionReminderNotification{}
	NewPractitionerReminderSchedulerForTest(practitionerRepo, sessionReminders, f.svc.crypto,
		WithPractitionerReminderWindow(f.start.Sub(time.Now())+time.Minute)).TickOnce(context.Background())
	require.Len(t, sessionReminders.calls, 1)
	assert.Equal(t, f.availability.UserID, sessionReminders.calls[0].PartnerID)
	assert.Equal(t, ports.BookingOriginPractitioner, sessionReminders.calls[0].Origin)
}

func TestBuildNotificationData_CarriesTheOrigin(t *testing.T) {
	svc := &BookingService{}
	clientID := uuid.New()
	online := &domain.Booking{ClientID: &clientID, Origin: domain.BookingOriginOnline}
	practitioner := &domain.Booking{ClientID: &clientID, Origin: domain.BookingOriginPractitioner}

	assert.Equal(t, ports.BookingOriginOnline, svc.buildNotificationData(online, "").Origin)
	assert.Equal(t, ports.BookingOriginPractitioner, svc.buildNotificationData(practitioner, "").Origin)
}
