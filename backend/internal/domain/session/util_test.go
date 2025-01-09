package sessionService_test

import (
	"time"

	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/tests"
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
