package types

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventFormat = "2006-01-02"
	EventPrice  = 4000 // the price for one ticket for an event
)

// got from the admin
type EventForm struct {
	Location        string    `json:"location"`
	PlaceCount      int       `json:"placecount"`
	BeginAt         time.Time `json:"beginat"`
	SessionDuration time.Time `json:"sessionduration"`
	Day             int       `json:"day"`
	Month           int       `json:"month"`
	Year            int       `json:"year"`
}

// the db version
type Event struct {
	Id              string        `json:"id"`
	Location        string        `json:"location"`
	PlaceCount      int           `json:"placecount"`
	BeginAt         time.Time     `json:"beginat"`
	SessionDuration time.Duration `json:"sessionduration"`
	PriceId         string        `json:"priceid"`
	Day             int           `json:"day"`
	Month           int           `json:"month"`
	Year            int           `json:"year"`
}

// TODO: I need the price id ? It seems that I do not.
type EventSent struct {
	Id              string        `json:"id"`
	Location        string        `json:"location"`
	PlaceCount      int           `json:"placecount"`
	FreePlace       int           `json:"freeplace"`
	BeginAt         time.Time     `json:"beginat"`
	SessionDuration time.Duration `json:"sessionduration"`
	Day             int           `json:"day"`
	Month           int           `json:"month"`
	Year            int           `json:"year"`
}

type EventBody struct {
	PastEvents     []*EventSent `json:"pastEvents"`
	NextEvents     []*EventSent `json:"nextEvents"`
	IncomingEvents []*EventSent `json:"incomingEvents"`
}

func NewEvent(location string, placecount int, date time.Time, priceid string) *Event {
	return &Event{
		Id:              uuid.NewString(),
		Location:        location,
		BeginAt:         date,
		PlaceCount:      placecount,
		PriceId:         priceid,
		SessionDuration: 30 * time.Minute,
		Day:             1,
		Month:           1,
		Year:            1,
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
