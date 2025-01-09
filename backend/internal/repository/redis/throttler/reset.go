package throttlerRepository

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/GaryHY/leviosa/internal/domain/throttler"
	rp "github.com/GaryHY/leviosa/internal/repository"

	"github.com/redis/go-redis/v9"
)

// Reset is called with a successful sign in to reset the throtlting associated with the provided email.
// TODO:
// - do better error handling with that
// - use transaction
func (t *Repository) Reset(ctx context.Context, key string) error {
	var info throttlerService.Info
	val, err := t.client.Get(ctx, THROTTLERPREFIX+key).Bytes()
	if err != nil {
		switch {
		case errors.Is(err, redis.ErrClosed), errors.As(err, &net.OpError{}):
			return rp.NewDatabaseErr(err)
		case errors.Is(err, redis.Nil):
			return rp.NewNotFoundErr(err, "")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	json.Unmarshal(val, &info)
	info.Email = key
	info.Attempts = 0
	info.LastAttempt = time.Time{}
	info.LockedUntil = time.Time{}
	encoded, err := json.Marshal(info)
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	err = t.client.Set(ctx, THROTTLERPREFIX+key, encoded, throttlerService.THROTTLERSESSIONDURATION).Err()
	if err != nil {
		switch {
		case errors.Is(err, redis.ErrClosed), errors.As(err, &net.OpError{}):
			return rp.NewDatabaseErr(err)
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	return nil
}
