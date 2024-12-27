package throttlerRepository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	throttlerService "github.com/GaryHY/event-reservation-app/internal/domain/throttler"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

// MakeAttempt is called when there is sign in error and adds an attempt for the provided email.
func (t *Repository) MakeAttempt(ctx context.Context, key string, now time.Time) error {
	var info throttlerService.Info
	// TODO: I need to make that thing better, there is a Exists function for that. Look for the other functions
	// like GetSet, GetDel etc..
	val, err := t.client.Get(ctx, THROTTLERPREFIX+key).Bytes()
	switch {

	case errors.Is(err, redis.Nil):
		// create an entry for that email
		info := throttlerService.NewInfo(key)
		err = t.client.Set(ctx, THROTTLERPREFIX+info.Email, info, throttlerService.THROTTLERSESSIONDURATION).Err()
		if err != nil {
			switch {
			case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
				return rp.NewContextErr(err)
				// case errors.Is(err, redis.ErrClosed), errors.As(err, &net.OpError{}):
			case errors.Is(err, redis.ErrClosed):
				fallthrough
			default:
				return rp.NewDatabaseErr(err)
			}
		}
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		fallthrough
	default:
		return rp.NewDatabaseErr(err)
	}

	json.Unmarshal(val, &info)
	info.Email = key
	info.Attempts++
	info.LastAttempt = now
	// handle the case when the maxattempts in hit
	if info.Attempts >= throttlerService.MAXATTEMPT {
		info.LockedUntil = now.Add(throttlerService.THROTTLINGDURATION)
		info.Attempts = 0 // Reset attempts after lock
	}
	encoded, err := json.Marshal(info)
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if err = t.client.Set(ctx, THROTTLERPREFIX+info.Email, encoded, throttlerService.THROTTLERSESSIONDURATION).Err(); err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		case errors.Is(err, redis.ErrClosed):
			return rp.NewDatabaseErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	return nil
}

func setNewInfo(ctx context.Context, client *redis.Client, email string, encoded []byte) error {
	err := client.Set(ctx, THROTTLERPREFIX+email, encoded, throttlerService.THROTTLERSESSIONDURATION).Err()
	if err != nil {
		switch {
		case errors.Is(err, redis.ErrClosed):
			return rp.NewDatabaseErr(err)
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	return nil
}
