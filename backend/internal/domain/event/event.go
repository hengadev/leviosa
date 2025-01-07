package eventService

import (
	"context"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/stripe"
	"github.com/google/uuid"
)

// prohibitedFields precise the fields that are non updatable on the event is created
var prohibitedFields = []string{
	"ID",
	"PriceID",
}

type Event struct {
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	Description      string        `json:"description"`
	Type             string        `json:"type"`
	Location         string        `json:"location"`
	PlaceCount       int           `json:"place_count"`
	FreePlace        int           `json:"free_place"`
	BeginAt          time.Time     `json:"beginat"`
	BeginAtFormatted string        `json:"beginat_formatted"`
	EndAt            time.Time     `json:"end_at"`
	EndAtFormatted   string        `json:"end_at_formatted"`
	SessionDuration  time.Duration `json:"session_duration"`
	PriceID          string        `json:"-"`
	Day              int           `json:"day"`
	Month            int           `json:"month"`
	Year             int           `json:"year"`
}

func NewEvent(
	location string,
	placecount int,
	beginat time.Time,
	sessionduration time.Duration,
	day int,
	month int,
	year int,
) *Event {
	return &Event{
		ID:              uuid.NewString(),
		Location:        location,
		BeginAt:         beginat,
		SessionDuration: sessionduration,
		PlaceCount:      placecount,
		FreePlace:       placecount,
		Day:             day,
		Month:           month,
		Year:            year,
	}
}

const description = "1 X Pass valuable for all the event."
const price = 1999

func (e *Event) getProductName() string {
	return fmt.Sprintf("Ticket pour l'evenement du : %s", e.SessionDuration)
}

func (e *Event) GetPaymentInfo(ctx context.Context) *stripeService.PaymentProductInfo {
	return &stripeService.PaymentProductInfo{
		ID:          e.ID,
		Name:        e.getProductName(),
		Description: description,
		Price:       price,
	}
}
