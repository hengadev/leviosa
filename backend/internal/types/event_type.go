package types

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventFormat = "2006-01-02"
	EventPrice  = 4000 // the price for one ticket for an event
)

type EventForm struct {
	Location   string    `json:"location"`
	PlaceCount int       `json:"placecount"`
	Date       time.Time `json:"date"`
}

type Event struct {
	Id         string    `json:"id"`
	Location   string    `json:"location"`
	PlaceCount int       `json:"placecount"`
	Date       time.Time `json:"date"`
	PriceId    string    `json:"priceid"`
}

func NewEvent(location string, placecount int, date time.Time, priceid string) *Event {
	return &Event{
		Id:         uuid.NewString(),
		Location:   location,
		Date:       date,
		PlaceCount: placecount,
		PriceId:    priceid,
	}
}

// NOTE: old function
// func NewEvent(placecount int) *Event {
// 	return &Event{
// 		Id:       uuid.NewString(),
// 		Location: "Some Location",
//      Date:       time.Now().Format(EventFormat),
// 		PlaceCount: placecount,
// 	}
// }
