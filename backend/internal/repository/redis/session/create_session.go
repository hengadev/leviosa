package sessionRepository

import (
	"context"
	"errors"
	"net"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

func (s *Repository) CreateSession(ctx context.Context, sessionID string, sessionEncoded []byte) error {
	err := s.client.Set(ctx, SESSIONPREFIX+sessionID, sessionEncoded, sessionService.SessionDuration).Err()
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
