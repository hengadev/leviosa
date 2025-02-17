package eventService

import (
	"context"
	"fmt"
	"time"

	"github.com/GaryHY/leviosa/internal/domain/stripe"
	"github.com/google/uuid"
)

type Event struct {
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	Description      string        `json:"description"`
	Type             string        `json:"type"`
	Location         string        `json:"location"`
	PlaceCount       int           `json:"place_count"`
	FreePlace        int           `json:"free_place"`
	BeginAt          time.Time     `json:"begin_at"`
	BeginAtFormatted string        `json:"begin_at_formatted"`
	EndAt            time.Time     `json:"end_at"`
	EndAtFormatted   string        `json:"end_at_formatted"`
	SessionDuration  time.Duration `json:"session_duration"`
	PriceID          string        `json:"-"`
	Day              int           `json:"day"`
	Month            int           `json:"month"`
	Year             int           `json:"year"`
}

type EventBis struct {
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	Description      string        `json:"description"`
	Type             string        `json:"type"`
	City             string        `json:"city"`
	PostalCode       string        `json:"postal_code"`
	Address1         string        `json:"address1"`
	Address2         string        `json:"address2"`
	PlaceCount       int           `json:"place_count"`
	FreePlace        int           `json:"free_place"`
	BeginAt          time.Time     `json:"begin_at"`
	BeginAtFormatted string        `json:"begin_at_formatted"`
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

func (e Event) AssertComparable() {}

func (e Event) GetSQLColumnMapping() map[string]string {
	return map[string]string{
		"ID":               "id",
		"Title":            "title",
		"Description":      "description",
		"Type":             "type",
		"Location":         "location",
		"PlaceCount":       "placecount",
		"FreePlace":        "freeplace",
		"BeginAtFormatted": "begin_at_formatted",
		"EndAtFormatted":   "end_at_formatted",
		"SessionDuration":  "session_duration",
		"PriceID":          "price_id",
		"Day":              "day",
		"Month":            "month",
		"Year":             "year",
	}
}

func (e Event) GetProhibitedFields() []string {
	return []string{"ID", "PriceID"}
}
