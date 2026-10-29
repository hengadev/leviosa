package booking

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
)

// Payment hold expiry (ADR-0018): bookings still awaiting payment once their
// hold elapsed become expired, and their PaymentIntents are canceled so the
// client can no longer pay for a slot that is offered again. A payment that
// still goes through is handled as a late payment (late_payment.go).

// ExpireElapsedHolds expires the bookings past their payment hold in one
// conditional update (a booking confirmed meanwhile is left alone, and two
// instances never expire the same booking twice), then cancels their
// PaymentIntents. A cancel that fails is logged: Stripe refuses canceling an
// intent that succeeded meanwhile, and its succeeded webhook settles it.
func (s *BookingService) ExpireElapsedHolds(ctx context.Context) (int, error) {
	expired, err := s.bookingRepo.ExpireElapsedHolds(ctx, time.Now())
	if err != nil {
		return 0, fmt.Errorf("expire elapsed payment holds: %w", err)
	}
	for _, hold := range expired {
		slog.InfoContext(ctx, "payment hold expired", "booking_id", hold.BookingID)
		if hold.PaymentIntentID == nil || *hold.PaymentIntentID == "" {
			continue
		}
		if err := s.paymentService.CancelPaymentIntent(ctx, *hold.PaymentIntentID); err != nil {
			slog.WarnContext(ctx, "failed to cancel the payment intent of an expired hold",
				"booking_id", hold.BookingID, "error", err)
		}
	}
	return len(expired), nil
}

// holdExpirer is what the job needs from the booking service.
type holdExpirer interface {
	ExpireElapsedHolds(ctx context.Context) (int, error)
}

// DefaultHoldExpiryInterval is how often the job looks for elapsed holds.
// Slots are freed at read time as soon as a hold elapses (OccupiesTime);
// the job only settles the status and cancels the intent.
const DefaultHoldExpiryInterval = time.Minute

// HoldExpiryJob is an in-process background worker that periodically expires
// elapsed payment holds. It follows the ReminderScheduler's lifecycle: Start
// blocks until its context is cancelled or Stop is called.
type HoldExpiryJob struct {
	expirer  holdExpirer
	interval time.Duration

	mu     sync.Mutex
	stopCh chan struct{}
}

// NewHoldExpiryJob creates the job; an interval <= 0 uses
// DefaultHoldExpiryInterval.
func NewHoldExpiryJob(service ports.BookingService, interval time.Duration) *HoldExpiryJob {
	return newHoldExpiryJob(service, interval)
}

func newHoldExpiryJob(expirer holdExpirer, interval time.Duration) *HoldExpiryJob {
	if interval <= 0 {
		interval = DefaultHoldExpiryInterval
	}
	return &HoldExpiryJob{expirer: expirer, interval: interval, stopCh: make(chan struct{})}
}

// Start runs the job: a first tick at once, then one every interval after
// the previous tick finished.
func (j *HoldExpiryJob) Start(ctx context.Context) {
	slog.InfoContext(ctx, "payment hold expiry job started", "interval", j.interval)

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "payment hold expiry job stopped")
			return
		case <-j.stopCh:
			slog.InfoContext(ctx, "payment hold expiry job stopped via Stop")
			return
		case <-timer.C:
			j.TickOnce(ctx)
			timer.Reset(j.interval)
		}
	}
}

// Stop signals the job to stop. Safe to call more than once, from any
// goroutine.
func (j *HoldExpiryJob) Stop() {
	j.mu.Lock()
	defer j.mu.Unlock()
	select {
	case <-j.stopCh:
	default:
		close(j.stopCh)
	}
}

// TickOnce runs one expiry pass. Exposed for tests.
func (j *HoldExpiryJob) TickOnce(ctx context.Context) {
	if j.expirer == nil {
		return
	}
	if _, err := j.expirer.ExpireElapsedHolds(ctx); err != nil {
		slog.ErrorContext(ctx, "payment hold expiry job: tick failed", "error", err)
	}
}
