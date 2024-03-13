package types

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventFormat = "2006-01-02"
	EventPrice  = 4000 // the price for one ticket for an event
)

type EventDate struct {
	Day   string `json:"day"`
	Month string
	Year  string
}

type Event struct {
	Id         string `json:"id"`
	Location   string `json:"location"`
	PlaceCount int    `json:"placecount"`
	Date       string `json:"date"`
	// TODO: Put the date in a time.Time ? Just need to parse the string parsed in the database with the eventFormat ?
}

// TODO: Finish them things
func NewEvent(placecount int) *Event {
	return &Event{
		Id:         uuid.NewString(),
		Location:   "Some Location",
		Date:       time.Now().Format(EventFormat),
		PlaceCount: placecount,
	}
}
