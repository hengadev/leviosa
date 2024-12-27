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

// type Event struct {
// 	ID              string        `json:"id"`
// 	Location        string        `json:"location"`
// 	PlaceCount      int           `json:"placecount"`
// 	FreePlace       int           `json:"freeplace"`
// 	BeginAt         time.Time     `json:"beginat"`
// 	SessionDuration time.Duration `json:"sessionduration"`
// 	PriceID         string        `json:"-"`
// 	Day             int           `json:"day"`
// 	Month           int           `json:"month"`
// 	Year            int           `json:"year"`
// }

// NOTE: type for event :
// - Leviosa Meetup
// - Leviosa Care
// - Leviosa Mental health

type Event struct {
	ID              string        `json:"id"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	Type            int           `json:"type"`
	Location        string        `json:"location"`
	PlaceCount      int           `json:"place_count"`
	FreePlace       int           `json:"free_place"`
	BeginAt         time.Time     `json:"beginat"`
	EndAt           time.Time     `json:"end_at"`
	SessionDuration time.Duration `json:"session_duration"`
	PriceID         string        `json:"-"`
	Day             int           `json:"day"`
	Month           int           `json:"month"`
	Year            int           `json:"year"`
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

type EventUser struct {
	PastEvents     []*Event `json:"pastEvents"`
	NextEvents     []*Event `json:"nextEvents"`
	IncomingEvents []*Event `json:"incomingEvents"`
}

func (e *Event) Create() {
	e.ID = uuid.NewString()
	// TODO: do the price id thing so I can actually acces it through stripe
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
