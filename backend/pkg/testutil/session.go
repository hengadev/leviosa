package testutil

import (
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
)

const SessionID = "a0rg34tWfQ33009_K"

var sessionTime, _ = time.Parse(time.Layout, "07/12 11:00:00AM '98 -0700")

var BaseSession = sessionService.Session{
	ID:         SessionID,
	UserID:     1,
	Role:       "basic",
	LoggedInAt: sessionTime,
	CreatedAt:  sessionTime,
	ExpiresAt:  sessionTime.Add(time.Hour),
}

var InitSession = map[string]*sessionService.Values{
	BaseSession.ID: BaseSession.Values(),
}
