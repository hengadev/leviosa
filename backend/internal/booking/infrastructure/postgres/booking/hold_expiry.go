package bookingRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// Payment hold expiry and late payments (ADR-0018). Every transition here is
// a conditional update on the booking's status, so the expiry job, a payment
// webhook and its duplicate deliveries never overwrite one another.

// ConfirmGuarded takes the partner's agenda lock (as CreateGuarded does),
// locks the booking row, and confirms it only if verify accepts the
// partner's other occupying bookings as they stand under the lock.
func (r *Repository) ConfirmGuarded(ctx context.Context, id, partnerID uuid.UUID, verify func(current *domain.BookingEncx, others []*domain.BookingEncx) error) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, errs.ClassifyPgError("begin guarded booking confirmation", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // a no-op once committed

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1::text, 0))`, partnerLockKey(partnerID)); err != nil {
		return false, errs.ClassifyPgError("lock partner agenda", err)
	}

	lock := fmt.Sprintf(`SELECT id FROM %s.bookings WHERE id = $1 FOR UPDATE`, r.schema)
	var lockedID uuid.UUID
	if err := tx.QueryRow(ctx, lock, id).Scan(&lockedID); err != nil {
		return false, errs.ClassifyPgError("lock booking", err)
	}

	current, err := r.selectBookings(ctx, tx, "get booking to confirm", `WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	if len(current) != 1 {
		return false, errs.ErrRepositoryNotFound
	}
	if current[0].Status != domain.BookingStatusAwaitingPayment && current[0].Status != domain.BookingStatusExpired {
		return false, nil
	}

	occupying, err := r.occupyingBookingsByPartner(ctx, tx, partnerID)
	if err != nil {
		return false, err
	}
	others := make([]*domain.BookingEncx, 0, len(occupying))
	for _, b := range occupying {
		if b.ID != id {
			others = append(others, b)
		}
	}
	if err := verify(current[0], others); err != nil {
		return false, err
	}

	update := fmt.Sprintf(`
		UPDATE %s.bookings
		SET status = $2, payment_status = $3, updated_at = NOW()
		WHERE id = $1
	`, r.schema)
	if _, err := tx.Exec(ctx, update, id, domain.BookingStatusConfirmed, domain.PaymentStatusPaid); err != nil {
		return false, errs.ClassifyPgError("confirm booking", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, errs.ClassifyPgError("commit guarded booking confirmation", err)
	}
	return true, nil
}

// ExpireElapsedHolds is one UPDATE ... RETURNING: a row is moved by exactly
// one caller, and a booking confirmed first (its row locked by
// ConfirmGuarded) no longer matches once the lock is released.
func (r *Repository) ExpireElapsedHolds(ctx context.Context, now time.Time) ([]ports.ExpiredHold, error) {
	query := fmt.Sprintf(`
		UPDATE %s.bookings
		SET status = $1, updated_at = NOW()
		WHERE status = $2 AND hold_expires_at IS NOT NULL AND hold_expires_at <= $3
		RETURNING id, payment_intent_id
	`, r.schema)
	rows, err := r.pool.Query(ctx, query, domain.BookingStatusExpired, domain.BookingStatusAwaitingPayment, now)
	if err != nil {
		return nil, errs.ClassifyPgError("expire elapsed holds", err)
	}
	defer rows.Close()

	expired := []ports.ExpiredHold{}
	for rows.Next() {
		var hold ports.ExpiredHold
		if err := rows.Scan(&hold.BookingID, &hold.PaymentIntentID); err != nil {
			return nil, errs.ClassifyPgError("scan expired hold", err)
		}
		expired = append(expired, hold)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ClassifyPgError("iterate expired holds", err)
	}
	return expired, nil
}

func (r *Repository) ExpireHold(ctx context.Context, id uuid.UUID, now time.Time) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s.bookings
		SET status = $2, updated_at = NOW()
		WHERE id = $1 AND status = $3 AND hold_expires_at IS NOT NULL AND hold_expires_at <= $4
	`, r.schema)
	result, err := r.pool.Exec(ctx, query, id, domain.BookingStatusExpired, domain.BookingStatusAwaitingPayment, now)
	if err != nil {
		return false, errs.ClassifyPgError("expire hold", err)
	}
	return result.RowsAffected() == 1, nil
}

func (r *Repository) ClaimLatePaymentRefund(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s.bookings
		SET payment_status = $2, updated_at = NOW()
		WHERE id = $1 AND status = $3 AND payment_status <> $2
	`, r.schema)
	result, err := r.pool.Exec(ctx, query, id, domain.PaymentStatusRefunded, domain.BookingStatusExpired)
	if err != nil {
		return false, errs.ClassifyPgError("claim late payment refund", err)
	}
	return result.RowsAffected() == 1, nil
}

func (r *Repository) ReleaseLatePaymentRefund(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`
		UPDATE %s.bookings
		SET payment_status = $2, updated_at = NOW()
		WHERE id = $1 AND status = $3 AND payment_status = $4
	`, r.schema)
	if _, err := r.pool.Exec(ctx, query, id, domain.PaymentStatusPaid, domain.BookingStatusExpired, domain.PaymentStatusRefunded); err != nil {
		return errs.ClassifyPgError("release late payment refund", err)
	}
	return nil
}
