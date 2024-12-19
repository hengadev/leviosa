package testutil

import (
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

const SessionID = "a36bde82-55ae-4fa9-9289-cb82bde68014"
const RandomSessionID = "3e3e6273-f118-4259-89a5-abb89cdd7492"

// var sessionTime, _ = time.Parse(time.Layout, "07/12 11:00:00AM '98 -0700")

var BaseSession = sessionService.Session{
	ID:        SessionID,
	UserID:    "user123",
	Role:      userService.BASIC,
	ExpiresAt: time.Now().Add(15 * time.Minute),
}

var InitSession = map[string]*sessionService.Values{
	BaseSession.ID: BaseSession.Values(),
}
