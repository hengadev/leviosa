package redis

import (
	"context"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	Client *redis.Client
}

func NewSessionRepository(ctx context.Context, client *redis.Client) (*SessionRepository, error) {
	return &SessionRepository{client}, nil
}

func (s *SessionRepository) RemoveSession(ctx context.Context, sessionID string) error {
	if err := s.Client.Del(ctx, sessionID).Err(); err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}
