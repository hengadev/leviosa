package booking

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
	"github.com/hengadev/encx"
)

// practitionerReminderRepo is the subset of BookingRepository that the
// practitioner session-reminder scheduler needs (issue 046).
type practitionerReminderRepo interface {
	FindBookingsDueForPractitionerReminder(ctx context.Context) ([]*domain.BookingEncx, error)
	MarkPractitionerReminderSent(ctx context.Context, bookingID uuid.UUID) error
	ClaimPractitionerReminder(ctx context.Context, bookingID uuid.UUID) (bool, error)
}

// PractitionerSessionReminderNotifier is the notification port the scheduler
// sends through: the booking adapter's Web Push path (ADR-0020).
type PractitionerSessionReminderNotifier interface {
	SendPractitionerSessionReminder(ctx context.Context, data ports.BookingNotificationData) error
}

// PractitionerReminderScheduler is an in-process background worker that
// periodically notifies the practitioner of sessions starting within the
// next 15 minutes ("Dans 15 min", issue 046). Only started when Web Push is
// configured (VAPID keys present).
type PractitionerReminderScheduler struct {
	repo         practitionerReminderRepo
	notification PractitionerSessionReminderNotifier
	crypto       encx.CryptoService

	interval       time.Duration // how often to tick (default: 1 minute)
	reminderWindow time.Duration // how far ahead to look (default: 15 minutes)

	mu     sync.Mutex
	stopCh chan struct{}
}

// NewPractitionerReminderScheduler creates a new scheduler. Accepts the full
// ports.BookingRepository (which satisfies practitionerReminderRepo).
func NewPractitionerReminderScheduler(
	bookingRepo ports.BookingRepository,
	notification PractitionerSessionReminderNotifier,
	crypto encx.CryptoService,
	opts ...PractitionerSchedulerOption,
) *PractitionerReminderScheduler {
	return newPractitionerReminderScheduler(bookingRepo, notification, crypto, opts...)
}

// NewPractitionerReminderSchedulerForTest creates a scheduler with a
// narrower repository interface suitable for unit testing.
func NewPractitionerReminderSchedulerForTest(
	repo practitionerReminderRepo,
	notification PractitionerSessionReminderNotifier,
	crypto encx.CryptoService,
	opts ...PractitionerSchedulerOption,
) *PractitionerReminderScheduler {
	return newPractitionerReminderScheduler(repo, notification, crypto, opts...)
}

func newPractitionerReminderScheduler(
	repo practitionerReminderRepo,
	notification PractitionerSessionReminderNotifier,
	crypto encx.CryptoService,
	opts ...PractitionerSchedulerOption,
) *PractitionerReminderScheduler {
	s := &PractitionerReminderScheduler{
		repo:           repo,
		notification:   notification,
		crypto:         crypto,
		interval:       time.Minute,
		reminderWindow: 15 * time.Minute,
		stopCh:         make(chan struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// PractitionerSchedulerOption configures optional scheduler settings.
type PractitionerSchedulerOption func(*PractitionerReminderScheduler)

// WithPractitionerInterval sets the tick interval. Must be > 0.
func WithPractitionerInterval(d time.Duration) PractitionerSchedulerOption {
	return func(s *PractitionerReminderScheduler) {
		if d > 0 {
			s.interval = d
		}
	}
}

// WithPractitionerReminderWindow sets how far ahead the scheduler looks for
// upcoming sessions. Must be > 0.
func WithPractitionerReminderWindow(d time.Duration) PractitionerSchedulerOption {
	return func(s *PractitionerReminderScheduler) {
		if d > 0 {
			s.reminderWindow = d
		}
	}
}

// Start begins the scheduler loop. It blocks until the context is cancelled
// or Stop is called. Non-blocking like ReminderScheduler.Start: a slow tick
// is skipped, not queued.
func (s *PractitionerReminderScheduler) Start(ctx context.Context) {
	slog.InfoContext(ctx, "practitioner session reminder scheduler started",
		"interval", s.interval,
		"reminder_window", s.reminderWindow,
	)

	timer := time.NewTimer(0) // fire immediately for first tick
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "practitioner session reminder scheduler stopped")
			return
		case <-s.stopCh:
			slog.InfoContext(ctx, "practitioner session reminder scheduler stopped via Stop")
			return
		case <-timer.C:
			s.tick(ctx)
			timer.Reset(s.interval)
		}
	}
}

// Stop signals the scheduler to stop. Safe from any goroutine.
func (s *PractitionerReminderScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
}

// TickOnce runs a single scheduler tick. Exposed for testing.
func (s *PractitionerReminderScheduler) TickOnce(ctx context.Context) {
	s.tick(ctx)
}

// tick runs one scheduler cycle: find confirmed, not-yet-reminded bookings
// starting within the reminder window, claim each reminder, and send the
// ones claimed. slot_start_time is encrypted, so time filtering happens here
// after decryption.
func (s *PractitionerReminderScheduler) tick(ctx context.Context) {
	if s.repo == nil {
		return
	}
	now := time.Now()
	windowEnd := now.Add(s.reminderWindow)

	bookingsEncx, err := s.repo.FindBookingsDueForPractitionerReminder(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "practitioner reminder scheduler: failed to query bookings", "error", err)
		return
	}

	for _, bx := range bookingsEncx {
		booking, err := domain.DecryptBookingEncx(ctx, s.crypto, bx)
		if err != nil {
			slog.ErrorContext(ctx, "practitioner reminder scheduler: failed to decrypt booking",
				"booking_id", bx.ID,
				"error", err,
			)
			continue
		}

		// Sessions already started: mark as reminded without sending so they
		// don't reappear on every tick.
		if !booking.SlotStartTime.After(now) {
			if err := s.repo.MarkPractitionerReminderSent(ctx, booking.ID); err != nil {
				slog.ErrorContext(ctx, "practitioner reminder scheduler: failed to mark started booking as reminded",
					"booking_id", booking.ID,
					"error", err,
				)
			}
			continue
		}

		// Sessions beyond the reminder window: skip without marking.
		if booking.SlotStartTime.After(windowEnd) {
			continue
		}

		// Claim the reminder before sending it: the conditional update only
		// succeeds while the booking is still confirmed and not yet reminded,
		// so the notification is sent at most once across ticks, restarts and
		// instances. The claim is kept even if sending fails.
		claimed, err := s.repo.ClaimPractitionerReminder(ctx, booking.ID)
		if err != nil {
			slog.ErrorContext(ctx, "practitioner reminder scheduler: failed to claim reminder",
				"booking_id", booking.ID,
				"error", err,
			)
			continue
		}
		if !claimed {
			continue
		}

		data := s.buildNotificationData(booking)
		if err := s.notification.SendPractitionerSessionReminder(ctx, data); err != nil {
			slog.ErrorContext(ctx, "practitioner reminder scheduler: failed to send reminder",
				"booking_id", booking.ID,
				"error", err,
			)
		}
	}

	slog.DebugContext(ctx, "practitioner reminder scheduler tick completed",
		"candidates", len(bookingsEncx),
		"elapsed", time.Since(now),
	)
}

// buildNotificationData creates BookingNotificationData from a decrypted
// booking. The reminder is for the practitioner, so it is sent whatever the
// booking's origin.
func (s *PractitionerReminderScheduler) buildNotificationData(booking *domain.Booking) ports.BookingNotificationData {
	data := ports.BookingNotificationData{
		BookingID:       booking.ID,
		PartnerID:       booking.PartnerID,
		RoomID:          booking.RoomID,
		ProductID:       booking.ProductID,
		SlotStartTime:   booking.SlotStartTime,
		SlotEndTime:     booking.SlotEndTime,
		TotalPriceCents: booking.TotalPriceCents,
		Currency:        booking.Currency,
		Locale:          string(booking.Locale),
		Origin:          string(booking.Origin),
	}

	if booking.ClientID != nil {
		data.ClientID = *booking.ClientID
	}

	if booking.IsGuestBooking() {
		data.IsGuestBooking = true
		data.GuestEmail = booking.GuestEmail
		data.GuestPhone = booking.GuestPhone
		data.ClientName = booking.GuestDisplayName()
		data.ClientFirstName = booking.GuestFirstName
		data.ClientLastName = booking.GuestLastName
		data.ClientEmail = booking.GuestEmail
		data.ClientPhone = booking.GuestPhone
	}

	data.Token = booking.Token

	return data
}
