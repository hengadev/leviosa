package session_repo

import (
	"context"
	"encoding/json"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// a function use for authentication and authorization.
func (s *SessionRepository) GetSessionIDByUserID(ctx context.Context, userID string) (string, error) {
	value, err := s.Client.Get(ctx, userID).Result()
	if err != nil {
		return "", rp.NewNotFoundError(err)
	}
	var sessionDecoded session.Session
	err = json.Unmarshal([]byte(value), &sessionDecoded)
	if err != nil {
		return "", err
	}
	return sessionDecoded.ID, nil
}
