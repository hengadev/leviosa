package sessionRepository

import (
	"context"
	"errors"

	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/redis/go-redis/v9"
)

func (s *Repository) FindSessionByID(ctx context.Context, sessionID string) ([]byte, error) {
	val, err := s.client.Get(ctx, SESSIONPREFIX+sessionID).Bytes()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, rp.NewNotFoundErr(err, "session")
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, rp.ErrContext
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return val, nil
}
