package session

import (
	"context"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"

	"github.com/google/uuid"
)

const SessionExpirationDuration = time.Minute * 20

type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userid"`
	Role       string    `json:"userrole"`
	LoggedInAt time.Time `json:"loggedinat"`
	CreatedAt  time.Time `json:"createdat"`
	ExpiresAt  time.Time `json:"expiresat"`
}

func NewSession(userID, role string) (*Session, error) {
	if err := uuid.Validate(userID); err != nil {
		return nil, err
	}
	return &Session{
		UserID: userID,
		Role:   role,
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

func (s *Session) Valid(ctx context.Context, minRole user.Role) (problems map[string]string) {
	var pbms = make(map[string]string)
	if err := uuid.Validate(s.ID); err != nil {
		pbms["id"] = "session ID is not of type UUID"
	}
	if err := uuid.Validate(s.UserID); err != nil {
		pbms["userid"] = "user ID is not of type UUID"
	}
	if time.Now().Add(SessionExpirationDuration).Before(s.ExpiresAt) {
		pbms["expiredat"] = "session expired"
	}
	sessionRole := user.ConvertToRole(s.Role)
	if !sessionRole.IsSuperior(minRole) {
		pbms["role"] = "unauthorized, user does not have the right priviledge"
	}
	return pbms
}
