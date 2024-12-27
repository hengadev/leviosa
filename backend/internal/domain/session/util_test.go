package sessionService_test

import (
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/tests"
)

var baseSession = &sessionService.Session{
	ID:        test.GenerateRandomString(16),
	UserID:    "1",
	Role:      models.BASIC,
	ExpiresAt: time.Now().Add(sessionService.SessionDuration),
}

var initMap = map[string]*sessionService.Values{
	baseSession.ID: baseSession.Values(),
}
