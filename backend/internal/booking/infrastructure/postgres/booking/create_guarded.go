package bookingRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// execer and querier are what the pool and a transaction have in common.
type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// CreateGuarded inserts a booking while holding a row lock on its
// availability. Slot times are encrypted at rest, so no database constraint
// can refuse an overlap: instead every booking creation on an availability
// takes the same lock, reads the bookings under it and lets verify decide.
// A second client submitting the same slot waits for the first one's commit
// and then sees its booking.
func (r *Repository) CreateGuarded(ctx context.Context, bookingEncx *domain.BookingEncx, verify func(existing []*domain.BookingEncx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return errs.ClassifyPgError("begin guarded booking creation", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // a no-op once committed

	lock := fmt.Sprintf(`SELECT id FROM %s.availabilities WHERE id = $1 FOR UPDATE`, r.schema)
	var lockedID any
	if err := tx.QueryRow(ctx, lock, bookingEncx.AvailabilityID).Scan(&lockedID); err != nil {
		return errs.ClassifyPgError("lock availability", err)
	}

	existing, err := r.bookingsByAvailability(ctx, tx, bookingEncx.AvailabilityID)
	if err != nil {
		return err
	}
	if err := verify(existing); err != nil {
		return err
	}

	if err := r.insert(ctx, tx, bookingEncx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return errs.ClassifyPgError("commit guarded booking creation", err)
	}
	return nil
}

// ConfirmAwaitingPayment is a conditional update: only the call that finds
// the booking still awaiting payment changes it, so concurrent deliveries of
// the same payment webhook confirm it once.
func (r *Repository) ConfirmAwaitingPayment(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s.bookings
		SET status = $2, payment_status = $3, updated_at = NOW()
		WHERE id = $1 AND status = $4
	`, r.schema)
	result, err := r.pool.Exec(ctx, query, id, domain.BookingStatusConfirmed, domain.PaymentStatusPaid, domain.BookingStatusAwaitingPayment)
	if err != nil {
		return false, errs.ClassifyPgError("confirm awaiting-payment booking", err)
	}
	return result.RowsAffected() == 1, nil
}
