package throttlerRepository

import (
	"context"
	"errors"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

// TODO: do the implementation for that function
func (t *Repository) IsLocked(ctx context.Context, key string) ([]byte, error) {
	val, err := t.client.Get(ctx, THROTTLERPREFIX+key).Bytes()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, rp.NewNotFoundErr(err, "throttler key")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return val, nil
}
