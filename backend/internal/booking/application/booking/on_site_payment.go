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

// RecordOnSitePayment records that the client paid the amount due at the
// practice, on the card terminal (see domain.RecordOnSitePayment).
func (s *BookingService) RecordOnSitePayment(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	return s.changeOnSitePayment(ctx, id, "record on-site payment", (*domain.Booking).RecordOnSitePayment)
}

// UndoOnSitePayment removes a recorded on-site payment, only on the
// Europe/Paris day it was recorded (see domain.UndoOnSitePayment).
func (s *BookingService) UndoOnSitePayment(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	return s.changeOnSitePayment(ctx, id, "undo on-site payment", (*domain.Booking).UndoOnSitePayment)
}

func (s *BookingService) changeOnSitePayment(ctx context.Context, id uuid.UUID, operation string, change func(*domain.Booking, time.Time) error) (*domain.Booking, error) {
	bookingEncx, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, errs.ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("get booking to %s: %w", operation, err)
	}

	booking, err := domain.DecryptBookingEncx(ctx, s.crypto, bookingEncx)
	if err != nil {
		return nil, fmt.Errorf("decrypt booking: %w", err)
	}

	if err := change(booking, s.now()); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	bookingEncx, err = domain.ProcessBookingEncx(ctx, s.crypto, booking)
	if err != nil {
		return nil, fmt.Errorf("encrypt booking: %w", err)
	}
	if err := s.bookingRepo.Update(ctx, bookingEncx); err != nil {
		return nil, fmt.Errorf("persist %s: %w", operation, err)
	}

	return booking, nil
}
