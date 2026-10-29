package domain

import (
	"time"

	"github.com/google/uuid"
)

// CreateAvailabilityExceptionRequest closes a recurring availability from
// start_date to end_date (YYYY-MM-DD, inclusive). Without end_date, only
// start_date is closed. There is deliberately no reason or label field
// (ADR-0019).
type CreateAvailabilityExceptionRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date,omitempty"`
}

// Dates parses the request into civil dates.
func (r CreateAvailabilityExceptionRequest) Dates() (start, end time.Time, err error) {
	start, err = ParseCivilDate(r.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if r.EndDate == "" {
		return start, start, nil
	}
	end, err = ParseCivilDate(r.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

type AvailabilityExceptionResponse struct {
	ID             uuid.UUID `json:"id"`
	AvailabilityID uuid.UUID `json:"availability_id"`
	StartDate      string    `json:"start_date"`
	EndDate        string    `json:"end_date"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateAvailabilityExceptionResponse reports the bookings already on the
// closed dates (confirmed, or awaiting payment with an unexpired hold). They
// are not changed.
type CreateAvailabilityExceptionResponse struct {
	Exception        AvailabilityExceptionResponse `json:"exception"`
	AffectedBookings int                           `json:"affected_bookings"`
}

func NewAvailabilityExceptionResponse(e AvailabilityException) AvailabilityExceptionResponse {
	return AvailabilityExceptionResponse{
		ID:             e.ID,
		AvailabilityID: e.AvailabilityID,
		StartDate:      e.StartDate.Format(time.DateOnly),
		EndDate:        e.EndDate.Format(time.DateOnly),
		CreatedAt:      e.CreatedAt,
	}
}
