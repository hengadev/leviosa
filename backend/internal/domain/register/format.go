package registerService

import (
	"context"
	"errors"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

// Format prepares the Registration struct for database storage by populating
// the BeginAtFormatted field with formatted value
// derived from the BeginAt field.
// This ensures consistency and proper formatting before persisting to the database.
func (r *Registration) Format(ctx context.Context) error {
	// TODO: check for the value to be rightfully formatted
	if r.StartTime.IsZero() {
		return domain.NewFormatError("registration StartTime field", errors.New("field is zero"))
	}
	r.StartTimeFormatted = r.StartTime.Format(time.RFC3339)
	r.StartTime = time.Time{}

	if r.EndTime.IsZero() {
		return domain.NewFormatError("registration EndTime field", errors.New("field is zero"))
	}
	r.EndTimeFormatted = r.EndTime.Format(time.RFC3339)
	r.EndTime = time.Time{}

	if r.CreatedAt.IsZero() {
		return domain.NewFormatError("registration CreatedAt field", errors.New("field is zero"))
	}
	r.CreatedAtFormatted = r.CreatedAt.Format(time.RFC3339)
	r.CreatedAt = time.Time{}

	if r.UpdatedAt.IsZero() {
		return domain.NewFormatError("registration UpdatedAt field", errors.New("field is zero"))
	}
	r.UpdatedAtFormatted = r.UpdatedAt.Format(time.RFC3339)
	r.UpdatedAt = time.Time{}
	return nil
}
