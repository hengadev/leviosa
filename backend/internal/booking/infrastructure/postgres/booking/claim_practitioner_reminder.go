package bookingRepository

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// ClaimPractitionerReminder sets practitioner_reminded_at to now() on a
// booking that is still confirmed and not yet practitioner-reminded, in one
// conditional update. It reports whether this call claimed the reminder:
// false when the booking was already reminded (by an earlier tick or another
// instance), is no longer confirmed (cancelled after the scheduler picked it
// up) or does not exist. Only the caller that claimed the reminder sends it,
// so each booking reminds the practitioner once.
func (r *Repository) ClaimPractitionerReminder(ctx context.Context, bookingID uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s.bookings SET practitioner_reminded_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'confirmed' AND practitioner_reminded_at IS NULL
	`, r.schema)

	result, err := r.pool.Exec(ctx, query, bookingID)
	if err != nil {
		return false, errs.ClassifyPgError("claim booking practitioner reminder", err)
	}
	return result.RowsAffected() == 1, nil
}
