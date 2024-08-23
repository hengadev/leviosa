package sessionRepository_test

import (
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
)

const sessionID = "a0rg34tWfQ33009_K"

var sessionTime, _ = time.Parse(time.Layout, "07/12 11:00:00AM '98 -0700")

var baseSession = sessionService.Session{
	ID:         sessionID,
	UserID:     1,
	Role:       "basic",
	LoggedInAt: sessionTime,
	CreatedAt:  sessionTime,
	ExpiresAt:  sessionTime.Add(time.Hour),
}

var initSession = map[string]any{
	baseSession.ID: baseSession.Values(),
}
