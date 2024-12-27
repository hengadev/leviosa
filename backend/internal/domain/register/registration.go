package register

import (
	"time"
)

// NOTE: you can register for :
// - an event
// - a classic consultation
// - a home consultation
// -> might need to add more in the future but if is just code

type Registration struct {
	UserID  string           `json:"userid"`
	Type    RegistrationType `json:"registration_type"`
	EventID string           `json:"eventid"`
	BeginAt time.Time        `json:"beginat"`
}

type EventRegistration struct {
	Registration
}

type RegistrationType uint

const (
	Consultation RegistrationType = iota
	Event
	AtHome
)

func NewRegistration(userID, eventID string, beginAt time.Time) *Registration {
	return &Registration{
		UserID:  userID,
		EventID: eventID,
		BeginAt: beginAt,
	}
}
