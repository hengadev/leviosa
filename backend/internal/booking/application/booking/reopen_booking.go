package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// ReopenBooking puts a completed or no-show booking back to confirmed,
// undoing a wrong close. Only allowed on the same Europe/Paris day the
// status was set (ADR-0006).
func (s *BookingService) ReopenBooking(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	// Get existing booking
	bookingEncx, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, errs.ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("get booking for reopen: %w", err)
	}

	// Decrypt booking
	booking, err := domain.DecryptBookingEncx(ctx, s.crypto, bookingEncx)
	if err != nil {
		return nil, fmt.Errorf("decrypt booking: %w", err)
	}

	// Reopen the booking
	if err := booking.Reopen(time.Now()); err != nil {
		return nil, fmt.Errorf("reopen booking: %w", err)
	}

	// Encrypt and persist changes
	bookingEncx, err = domain.ProcessBookingEncx(ctx, s.crypto, booking)
	if err != nil {
		return nil, fmt.Errorf("encrypt booking: %w", err)
	}

	if err := s.bookingRepo.Update(ctx, bookingEncx); err != nil {
		return nil, fmt.Errorf("update reopened booking: %w", err)
	}

	return booking, nil
}
