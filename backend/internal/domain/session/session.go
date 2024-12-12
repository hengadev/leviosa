package sessionService

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"

	"github.com/google/uuid"
)

const SessionDuration = 30 * 24 * time.Hour
const SessionName = "session_token"

type Session struct {
	ID         string           `json:"id"`
	UserID     string           `json:"userid"`
	Role       userService.Role `json:"userrole"`
	LoggedInAt time.Time        `json:"loggedinat"`
	CreatedAt  time.Time        `json:"createdat"`
	ExpiresAt  time.Time        `json:"expiresat"`
}

type Values struct {
	UserID     string           `json:"userid"`
	Role       userService.Role `json:"userrole"`
	LoggedInAt time.Time        `json:"loggedinat"`
	CreatedAt  time.Time        `json:"createdat"`
	ExpiresAt  time.Time        `json:"expiresat"`
}

func (s Session) IsZero() bool {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)
	vf := reflect.VisibleFields(t)
	for _, f := range vf {
		value := v.FieldByName(f.Name)
		if value.IsZero() && value.CanInterface() {
			return true
		}
	}
	return false
}

func (s *Session) Values() *Values {
	return &Values{
		UserID:     s.UserID,
		Role:       s.Role,
		LoggedInAt: s.LoggedInAt,
		CreatedAt:  s.CreatedAt,
		ExpiresAt:  s.ExpiresAt,
	}
}

func NewSession(userID string, role userService.Role) (*Session, error) {
	id := uuid.NewString()
	return &Session{
		ID:         id,
		UserID:     userID,
		Role:       role,
		LoggedInAt: time.Now(),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(SessionDuration),
	}, nil
}

func (s *Session) Valid(ctx context.Context, minRole userService.Role) (problems map[string]string) {
	var pbms = make(map[string]string)
	if err := uuid.Validate(s.ID); err != nil {
		pbms["id"] = "session ID is not of type UUID"
	}
	if time.Now().Add(SessionDuration).Before(s.ExpiresAt) {
		pbms["expiredat"] = "session expired"
	}
	if !s.Role.IsSuperior(minRole) {
		pbms["role"] = fmt.Sprintf("unauthorized, user role %s is not superior to %s", s.Role, minRole)
	}
	return pbms
}
