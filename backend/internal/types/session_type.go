package types

import (
	"github.com/google/uuid"
	"time"
)

// Create a new session, used in tests
func NewSession(user *User, user_id string) *Session {
	return &Session{
		Id:         uuid.NewString(),
		UserId:     user_id,
		Created_at: time.Now(),
	}
}

type Session struct {
	Id         string    `json:"id"`
	UserId     string    `json:"userid"`
	Created_at time.Time `json:"created_at"`
}

// A function to check whether a session is expired.
func (s *Session) isExpired() bool {
	return s.Created_at.Before(time.Now())
}
