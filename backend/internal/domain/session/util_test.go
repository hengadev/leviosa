package sessionService_test

import (
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/tests"
)

var baseSession = &sessionService.Session{
	ID:         test.GenerateRandomString(16),
	UserID:     1,
	Role:       userService.BASIC.String(),
	LoggedInAt: time.Now(),
	CreatedAt:  time.Now(),
	ExpiresAt:  time.Now().Add(sessionService.SessionDuration),
}

var initMap = map[string]*sessionService.Values{
	baseSession.ID: baseSession.Values(),
}
