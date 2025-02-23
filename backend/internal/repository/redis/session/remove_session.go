package sessionRepository

import (
	"context"
	"errors"
	"fmt"

	rp "github.com/GaryHY/leviosa/internal/repository"

	"github.com/redis/go-redis/v9"
)

func (s *Repository) RemoveSession(ctx context.Context, ID string) error {
	result := s.client.Del(ctx, SESSIONPREFIX+ID)

	if err := result.Err(); err != nil {
		switch {
		case errors.Is(err, redis.ErrClosed):
			return rp.NewDatabaseErr(err)
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	if result.Val() == 0 {
		return rp.NewNotFoundErr(fmt.Errorf("key does not exist"), "session")
	}
	return nil
}
