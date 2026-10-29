package ports

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/google/uuid"
)

// AvailabilityExceptionRepository persists availability exceptions (issue
// 012). Exceptions hold only dates, so they are stored in clear.
type AvailabilityExceptionRepository interface {
	// Create stores a new exception.
	Create(ctx context.Context, exception *domain.AvailabilityException) error

	// GetByID returns one exception, or errs.ErrRepositoryNotFound.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AvailabilityException, error)

	// ListByAvailabilityIDs returns the exceptions of the given
	// availabilities, ordered by start date. No ID yields no exception.
	ListByAvailabilityIDs(ctx context.Context, availabilityIDs []uuid.UUID) ([]domain.AvailabilityException, error)

	// Delete removes an exception, or returns errs.ErrRepositoryNotFound.
	Delete(ctx context.Context, id uuid.UUID) error
}
