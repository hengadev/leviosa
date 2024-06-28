package types

import (
	"time"
)

const (
	RegistrationDuration = time.Duration(30 * time.Minute)
)

// NOTE: The registration is for someone that wants to make a registration.
// NOTE: The BeginAt is because someone gets a creneau etc...
type Registration struct {
	EventId string    `json:"eventId"`
	UserId  string    `json:"userId"`
	BeginAt time.Time `json:"beginAt"`
	// IsPaid  bool   `json:"ispaid"`
	// TODO: Implement the isPaid thing depending on how I want to make the app.
}

type Validator struct {
	validate func(string) bool
}

func NewRegistration(eventId, userId string, beginAt time.Time) *Registration {
	return &Registration{
		EventId: eventId,
		UserId:  userId,
		BeginAt: beginAt,
	}
}
