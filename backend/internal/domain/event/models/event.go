package models

import (
	"context"
	"fmt"
	"time"

	"github.com/GaryHY/leviosa/internal/domain/stripe"

	"github.com/google/uuid"
)

type Event struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	City             string    `json:"city"`
	PostalCode       string    `json:"postal_code"`
	Address1         string    `json:"address1"`
	Address2         string    `json:"address2"`
	PlaceCount       int       `json:"place_count"`
	FreePlace        int       `json:"free_place"`
	BeginAt          time.Time `json:"begin_at"`
	EncryptedBeginAt string    `json:"begin_at_formatted"`
	EndAt            time.Time `json:"end_at"`
	EncryptedEndAt   string    `json:"end_at_formatted"`
	Products         []string  `json:"products"`
	Offers           []string  `json:"offers"`
	PriceID          string    `json:"-"`
	Day              int       `json:"day"`
	Month            int       `json:"month"`
	Year             int       `json:"year"`
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
		ID:         uuid.NewString(),
		City:       location,
		BeginAt:    beginat,
		PlaceCount: placecount,
		FreePlace:  placecount,
		Day:        day,
		Month:      month,
		Year:       year,
	}
}

const description = "1 X Pass valuable for all the event."
const price = 1999

func (e *Event) getProductName() string {
	return fmt.Sprintf("Ticket pour l'evenement : %s", e.Title)
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
