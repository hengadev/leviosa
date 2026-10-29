package ports

import (
	"context"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/google/uuid"
)

// AvailabilityExceptionService is the partner area's use of availability
// exceptions (issue 012). Every call names the calling partner's user ID: an
// availability of another partner answers errs.ErrRepositoryNotFound.
type AvailabilityExceptionService interface {
	// ListExceptions returns the exceptions of a recurring availability,
	// ordered by start date.
	ListExceptions(ctx context.Context, partnerUserID, availabilityID uuid.UUID) ([]domain.AvailabilityException, error)

	// CreateException closes [startDate, endDate] (civil dates, inclusive)
	// on a recurring availability. It returns the number of bookings of the
	// availability on the closed dates (confirmed, and awaiting payment with
	// an unexpired hold); it never changes them.
	CreateException(ctx context.Context, partnerUserID, availabilityID uuid.UUID, startDate, endDate time.Time) (*domain.AvailabilityException, int, error)

	// DeleteException reopens the dates of an exception.
	DeleteException(ctx context.Context, partnerUserID, availabilityID, exceptionID uuid.UUID) error
}
