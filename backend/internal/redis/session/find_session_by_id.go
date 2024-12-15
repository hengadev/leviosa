package sessionRepository

import (
	"context"
	"encoding/json"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Repository) FindSessionByID(ctx context.Context, sessionID string) (*sessionService.Session, error) {
	var res sessionService.Session
	val, err := s.Client.Get(ctx, SESSIONPREFIX+sessionID).Bytes()
	if err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	if err = json.Unmarshal(val, &res); err != nil {
		return nil, rp.NewInternalError(err)
	}
	res.ID = sessionID
	return &res, nil
}
