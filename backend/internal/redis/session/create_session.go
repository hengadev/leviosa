package sessionRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Repository) CreateSession(ctx context.Context, sessionID string, sessionEncoded []byte) error {
	err := s.client.Set(ctx, SESSIONPREFIX+sessionID, sessionEncoded, sessionService.SessionDuration).Err()
	if err != nil {
		return err
	}
	err = s.Client.Set(ctx, SESSIONPREFIX+session.ID, sessionEncoded, sessionService.SessionDuration).Err()
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}
