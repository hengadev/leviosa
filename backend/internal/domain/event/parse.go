package eventService

import (
	"time"

	"github.com/GaryHY/leviosa/internal/domain"
)

// Parse initializes the Event struct after retrieving data from the database
// by extracting and converting values from the BeginAtFormatted and EndAtFormatted
// fields into the BeginAt and EndAt fields.
// This ensures the Event struct is ready for application-level processing.
func (e *Event) Parse() error {
	beginAt, err := time.Parse(time.RFC3339, e.BeginAtFormatted)
	if err != nil {
		return domain.NewParsingError("registration StartTime field", err)
	}
	e.BeginAt = beginAt
	e.BeginAtFormatted = ""

	endAt, err := time.Parse(time.RFC3339, e.EndAtFormatted)
	if err != nil {
		return domain.NewParsingError("registration StartTime field", err)
	}
	e.EndAt = endAt
	e.EndAtFormatted = ""
	return nil
}
