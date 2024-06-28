package session

import (
	"time"

	"github.com/google/uuid"
)

const SessionExpirationDuration = time.Minute * 20

type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userid"`
	LoggedInAt time.Time `json:"loggedinat"`
	CreatedAt  time.Time `json:"createdat"`
	ExpiresAt  time.Time `json:"expiresat"`
}

func NewSession(userID string) (*Session, error) {
	if err := uuid.Validate(userID); err != nil {
		return nil, err
	}
	return &Session{
		UserID: userID,
	}, nil
}

// change the value of the field created at
func (s *Session) Create() {
	s.ID = uuid.NewString()
	s.CreatedAt = time.Now().UTC()
	s.ExpiresAt = time.Now().UTC().Add(SessionExpirationDuration)
}

func (s *Session) Login() {
	s.LoggedInAt = time.Now().UTC()
}

func (s *Session) Validate() map[string]string {
	// check all the field of that session and see if that thing is valid...
	var pbms = make(map[string]string)
	return pbms
}
