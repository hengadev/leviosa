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
	result := s.client.Set(ctx, SESSIONPREFIX+sessionID, sessionEncoded, sessionService.SessionDuration)
	if err := result.Err(); err != nil {
		switch {
		case errors.Is(err, redis.ErrClosed), errors.As(err, &net.OpError{}):
			return rp.NewDatabaseErr(err)
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}

	// TODO: Should I check for the insertion like so?
	if result.Val() == "" {
		return rp.NewNotCreatedErr(nil, "session")
	}
	return nil
}
