package booking

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"

	"github.com/google/uuid"
)

// WithAvailabilityExceptionRepository injects the availability exceptions
// (issue 012) the slot engine subtracts from recurring availabilities, for
// the public slot endpoint and booking creation alike. Without it no date is
// closed.
func WithAvailabilityExceptionRepository(r ports.AvailabilityExceptionRepository) Option {
	return func(s *BookingService) { s.exceptionRepo = r }
}

// exceptionsOf returns the exceptions of the given availabilities.
func (s *BookingService) exceptionsOf(ctx context.Context, availabilities ...*domain.Availability) ([]domain.AvailabilityException, error) {
	if s.exceptionRepo == nil {
		return nil, nil
	}
	var ids []uuid.UUID
	for _, availability := range availabilities {
		if availability.IsRecurring {
			ids = append(ids, availability.ID)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	exceptions, err := s.exceptionRepo.ListByAvailabilityIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list availability exceptions: %w", err)
	}
	return exceptions, nil
}
