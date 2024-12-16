package sessionService

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

func (s *Service) CreateSession(ctx context.Context, userID string, role userService.Role) (*Session, error) {
	session := NewSession(userID, role)
	sessionEncoded, err := json.Marshal(session)
	if err != nil {
		return "", app.NewJSONMarshalErr(err)
	}
	if err := s.Repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("create session in redis: %w", err)
	err = s.Repo.CreateSession(ctx, session.ID, sessionEncoded)
	}
	return session, nil
}
