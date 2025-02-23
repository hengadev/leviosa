package registerService

import (
	"time"

	"github.com/GaryHY/leviosa/internal/domain"
)

// Parse initializes the Registration struct after retrieving data from the database
// by extracting and converting values from the BeginAtFormatted
// field into the BeginAt and field.
// This ensures the Registration struct is ready for application-level processing.
func (r *Registration) Parse() error {
	startTime, err := time.Parse(time.RFC3339, r.StartTimeFormatted)
	if err != nil {
		return domain.NewParsingError("registration StartTime field", err)
	}
	r.StartTime = startTime
	r.StartTimeFormatted = ""

	endTime, err := time.Parse(time.RFC3339, r.EndTimeFormatted)
	if err != nil {
		return domain.NewParsingError("registration EndTime field", err)
	}
	r.EndTime = endTime
	r.EndTimeFormatted = ""

	createdAt, err := time.Parse(time.RFC3339, r.CreatedAtFormatted)
	if err != nil {
		return domain.NewParsingError("registration CreatedAt field", err)
	}
	r.CreatedAt = createdAt
	r.CreatedAtFormatted = ""

	updatedAt, err := time.Parse(time.RFC3339, r.UpdatedAtFormatted)
	if err != nil {
		return domain.NewParsingError("registration UpdatedAt field", err)
	}
	r.UpdatedAt = updatedAt
	r.UpdatedAtFormatted = ""
	return nil
}
