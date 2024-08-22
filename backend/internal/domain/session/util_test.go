package session_test

import (
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests"
)

var baseSession = &session.Session{
	ID:         test.GenerateRandomString(16),
	UserID:     1,
	Role:       userService.BASIC.String(),
	LoggedInAt: time.Now(),
	CreatedAt:  time.Now(),
	ExpiresAt:  time.Now().Add(session.SessionDuration),
}

var initMap = map[string]*session.Values{
	baseSession.ID: baseSession.Values(),
}
