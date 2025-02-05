package factories

import (
	"time"

	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
)

const SessionID = "a36bde82-55ae-4fa9-9289-cb82bde68014"
const RandomSessionID = "3e3e6273-f118-4259-89a5-abb89cdd7492"

// var sessionTime, _ = time.Parse(time.Layout, "07/12 11:00:00AM '98 -0700")

func NewBasicSession(overrides map[string]interface{}) *sessionService.Session {
	session := sessionService.Session{
		ID:        SessionID,
		UserID:    "user123",
		Role:      models.BASIC,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	for key, value := range overrides {
		switch key {
		case "ID":
			session.ID = value.(string)
		case "UserID":
			session.UserID = value.(string)
		case "Role":
			session.Role = value.(models.Role)
		case "ExpiresAt":
			session.ExpiresAt = value.(time.Time)
		}
	}
	return &session
}

var BaseSession = sessionService.Session{
	ID:        SessionID,
	UserID:    "user123",
	Role:      models.BASIC,
	ExpiresAt: time.Now().Add(15 * time.Minute),
}

var InitSession = map[string]*sessionService.Values{
	BaseSession.ID: BaseSession.Values(),
}
