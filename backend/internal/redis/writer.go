package redis

import (
	"context"
	"encoding/json"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// writer
func (s *SessionRepository) CreateSession(ctx context.Context, userSession *session.Session) (string, error) {
	sessionEncoded, err := json.Marshal(userSession)
	if err != nil {
		return "", err
	}
	err = s.Client.Set(ctx, userSession.ID, sessionEncoded, session.SessionDuration).Err()
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	return userSession.ID, nil
}

func (s *SessionRepository) RemoveSession(ctx context.Context, sessionID string) error {
	if err := s.Client.Del(ctx, sessionID).Err(); err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}
