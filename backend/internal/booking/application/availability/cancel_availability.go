package availability

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// CancelAvailability is how a partner deletes an availability, one-off or
// recurring (docs/issues/011-recurring-availabilities.md). The availability
// is deactivated, not removed: it stops offering slots at once, and its row
// stays so the bookings made against it keep their reference (the foreign
// key is ON DELETE RESTRICT). It never cancels, moves or otherwise changes an
// existing booking: those keep their time and status, and still occupy the
// partner's time until the practitioner cancels them one by one.
func (s *AvailabilityService) CancelAvailability(ctx context.Context, id uuid.UUID) error {
	// Get existing availability
	availabilityEncx, err := s.availabilityRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get availability for cancellation: %w", err)
	}

	availability, err := domain.DecryptAvailabilityEncx(ctx, s.crypto, availabilityEncx)
	if err != nil {
		return errs.NewNotDecryptedErr("availability", err)
	}

	// Cancel
	availability.Cancel()

	updatedAvailabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, s.crypto, availability)
	if err != nil {
		return errs.NewNotEncryptedErr("availability", err)
	}

	// Persist changes
	if err := s.availabilityRepo.Update(ctx, updatedAvailabilityEncx); err != nil {
		return fmt.Errorf("cancel availability: %w", err)
	}

	return nil
}
