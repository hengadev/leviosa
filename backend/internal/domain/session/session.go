package session

import (
	"context"
	"reflect"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"

	"github.com/google/uuid"
)

const SessionDuration = 30 * 24 * time.Hour
const SessionName = "session_token"

type Session struct {
	ID         string    `json:"id"`
	UserID     int       `json:"userid"`
	Role       string    `json:"userrole"`
	LoggedInAt time.Time `json:"loggedinat"`
	CreatedAt  time.Time `json:"createdat"`
	ExpiresAt  time.Time `json:"expiresat"`
}

type Values struct {
	UserID     int       `json:"userid"`
	Role       string    `json:"userrole"`
	LoggedInAt time.Time `json:"loggedinat"`
	CreatedAt  time.Time `json:"createdat"`
	ExpiresAt  time.Time `json:"expiresat"`
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

func NewSession(userID int, role string) (*Session, error) {
	return &Session{
		UserID: userID,
		Role:   role,
	}, nil
}

// change the value of the field created at
func (s *Session) Create() {
	s.ID = uuid.NewString()
	s.CreatedAt = time.Now().UTC()
	s.ExpiresAt = time.Now().UTC().Add(SessionDuration)
}

func (s *Session) Login() {
	s.LoggedInAt = time.Now().UTC()
}

func (s *Session) Valid(ctx context.Context, minRole userService.Role) (problems map[string]string) {
	var pbms = make(map[string]string)
	if err := uuid.Validate(s.ID); err != nil {
		pbms["id"] = "session ID is not of type UUID"
	}
	if time.Now().Add(SessionDuration).Before(s.ExpiresAt) {
		pbms["expiredat"] = "session expired"
	}
	sessionRole := userService.ConvertToRole(s.Role)
	if !sessionRole.IsSuperior(minRole) {
		pbms["role"] = "unauthorized, user does not have the right priviledge"
	}
	return pbms
}
