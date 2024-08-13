package session_test

import (
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests"
)

var sessionTest = &session.Session{
	ID:         test.GenerateRandomString(16),
	UserID:     1,
	Role:       user.BASIC.String(),
	LoggedInAt: time.Now(),
	CreatedAt:  time.Now(),
	ExpiresAt:  time.Now().Add(session.SessionDuration),
}

var initSessionValues = make(map[string]*session.Values)
