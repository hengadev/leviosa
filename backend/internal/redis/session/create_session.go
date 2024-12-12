package sessionRepository

import (
	"context"
	"encoding/json"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Repository) CreateSession(ctx context.Context, session *sessionService.Session) error {
	sessionEncoded, err := json.Marshal(session)
	if err != nil {
		return err
	}
	err = s.Client.Set(ctx, SESSIONPREFIX+session.ID, sessionEncoded, sessionService.SessionDuration).Err()
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}
