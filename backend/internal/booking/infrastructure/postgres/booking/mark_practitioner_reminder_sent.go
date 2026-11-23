package bookingRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// MarkPractitionerReminderSent sets practitioner_reminded_at to now() for a
// given booking, without sending anything. Used for bookings the scheduler
// will never remind (already started) so they stop being candidates.
func (r *Repository) MarkPractitionerReminderSent(ctx context.Context, bookingID uuid.UUID) error {
	query := fmt.Sprintf(`
		UPDATE %s.bookings SET practitioner_reminded_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, r.schema)

	result, err := r.pool.Exec(ctx, query, bookingID)
	if err != nil {
		return errs.ClassifyPgError("mark practitioner reminder sent", err)
	}

	if result.RowsAffected() == 0 {
		return errs.ErrRepositoryNotFound
	}

	return nil
}
