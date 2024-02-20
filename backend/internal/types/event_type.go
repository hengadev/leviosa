package types

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventFormat = "2006-01-02"
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
