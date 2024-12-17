package sessionRepository

import (
	"context"
	"errors"
	"net"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

func (s *Repository) RemoveSession(ctx context.Context, ID string) error {
	err := s.client.Del(ctx, SESSIONPREFIX+ID).Err()
	if err != nil {
		switch {
		case errors.Is(err, redis.ErrClosed), errors.As(err, &net.OpError{}):
			return rp.NewDatabaseErr(err)
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	return nil
}
