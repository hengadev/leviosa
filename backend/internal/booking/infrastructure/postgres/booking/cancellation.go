package bookingRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// Cancellation with a refund (ADR-0014). Like the hold-expiry transitions,
// each step is a conditional update on the booking's status, so two cancel
// requests at once refund once.

func (r *Repository) ClaimCancellation(ctx context.Context, id uuid.UUID, cancelledBy domain.CancellationActor, cancelledAt time.Time, refundCents int) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s.bookings
		SET status = $2, cancelled_at = $3, cancelled_by = $4, refunded_amount_cents = $5, updated_at = NOW()
		WHERE id = $1 AND status = $6
	`, r.schema)
	result, err := r.pool.Exec(ctx, query, id, domain.BookingStatusCancelled, cancelledAt, string(cancelledBy), refundCents, domain.BookingStatusConfirmed)
	if err != nil {
		return false, errs.ClassifyPgError("claim booking cancellation", err)
	}
	return result.RowsAffected() == 1, nil
}

func (r *Repository) ReleaseCancellation(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`
		UPDATE %s.bookings
		SET status = $2, cancelled_at = NULL, cancelled_by = NULL, refunded_amount_cents = NULL, updated_at = NOW()
		WHERE id = $1 AND status = $3
	`, r.schema)
	if _, err := r.pool.Exec(ctx, query, id, domain.BookingStatusConfirmed, domain.BookingStatusCancelled); err != nil {
		return errs.ClassifyPgError("release booking cancellation", err)
	}
	return nil
}

func (r *Repository) MarkCancellationRefunded(ctx context.Context, id uuid.UUID, paymentStatus domain.PaymentStatus) error {
	query := fmt.Sprintf(`
		UPDATE %s.bookings
		SET payment_status = $2, updated_at = NOW()
		WHERE id = $1 AND status = $3
	`, r.schema)
	if _, err := r.pool.Exec(ctx, query, id, paymentStatus, domain.BookingStatusCancelled); err != nil {
		return errs.ClassifyPgError("mark cancellation refunded", err)
	}
	return nil
}
