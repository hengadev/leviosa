package offerService

import (
	"context"
	"time"
)

// TODO: an offer can be of a type (massage, preparation mental etc..). These types are specified in the database by admin

type Offer struct {
	Name        string
	Duration    time.Time
	Description string
	Type        string // I get that thing from the database after initializing it
}

func NewOffer(name, description string, duration time.Time) *Offer {
	// TODO: validate the offer before creating it right ?
	return &Offer{
		Name:        name,
		Duration:    duration,
		Description: description,
	}
}

func (o Offer) Valid(ctx context.Context) map[string]string {
	return nil
}
