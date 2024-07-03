package session_repo

import (
	"context"
	"encoding/json"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *SessionRepository) CreateSession(ctx context.Context, userSession *session.Session) (string, error) {
	sessionEncoded, err := json.Marshal(userSession)
	if err != nil {
		return "", err
	}
	err = s.Client.Set(ctx, userSession.ID, sessionEncoded, session.SessionExpirationDuration).Err()
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	err = s.Client.Set(ctx, userSession.UserID, sessionEncoded, session.SessionExpirationDuration).Err()
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	return userSession.ID, nil
}
