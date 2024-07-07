package register

import (
	"time"
)

type Registration struct {
	UserID  string    `json:"userid"`
	EventID string    `json:"eventid"`
	BeginAt time.Time `json:"beginat"`
}

func NewRegistration(userID, eventID string, beginAt time.Time) *Registration {
	return &Registration{
		UserID:  userID,
		EventID: eventID,
		BeginAt: beginAt,
	}
}
