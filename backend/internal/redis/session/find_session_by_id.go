package sessionRepository

import (
	"context"
	"encoding/json"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Repository) FindSessionByID(ctx context.Context, sessionID string) (*session.Session, error) {
	var res session.Session
	val, err := s.Client.Get(ctx, sessionID).Bytes()
	if err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	json.Unmarshal(val, &res)
	res.ID = sessionID
	return &res, nil
}
