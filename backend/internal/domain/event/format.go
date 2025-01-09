package eventService

import (
	"context"
	"errors"
	"time"

	"github.com/GaryHY/leviosa/internal/domain"
)

// Format prepares the Event struct for database storage by populating
// the BeginAtFormatted and EndAtFormatted fields with formatted values
// derived from the BeginAt and EndAt fields.
// This ensures consistency and proper formatting before persisting to the database.
func (e *Event) Format(ctx context.Context) error {
	// TODO: use the same function than in registration
	if e.BeginAt.IsZero() {
		return domain.NewFormatError("registration StartTime field", errors.New("field is zero"))
	}
	e.BeginAtFormatted = e.BeginAt.Format(time.RFC3339)
	e.BeginAt = time.Time{}

	if e.EndAt.IsZero() {
		return domain.NewFormatError("registration StartTime field", errors.New("field is zero"))
	}
	e.EndAtFormatted = e.EndAt.Format(time.RFC3339)
	e.EndAt = time.Time{}
	return nil
}
