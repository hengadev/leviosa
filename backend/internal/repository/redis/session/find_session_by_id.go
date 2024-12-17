package sessionRepository

import (
	"context"
	"errors"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/redis/go-redis/v9"
)

func (s *Repository) FindSessionByID(ctx context.Context, sessionID string) ([]byte, error) {
	val, err := s.client.Get(ctx, SESSIONPREFIX+sessionID).Bytes()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, rp.NewNotFoundError(err, "session")
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			fallthrough
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return val, nil
}
