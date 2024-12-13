package throttlerRepository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/throttler"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/redis/go-redis/v9"
)

// MakeAttempt is called when there is sign in error and adds an attempt for the provided email.
func (t *Repository) MakeAttempt(ctx context.Context, key string, now time.Time) error {
	var info throttlerService.Info
	val, err := t.Client.Get(ctx, THROTTLERPREFIX+key).Bytes()
	switch {
	case err == redis.Nil:
		// create an entry for that email
		info := throttlerService.NewInfo(key)
		err = t.Client.Set(ctx, THROTTLERPREFIX+info.Email, info, throttlerService.THROTTLERSESSIONDURATION).Err()
		if err != nil {
			return rp.NewRessourceUpdateErr(err)
		}
		return nil
	case err != nil:
		return rp.NewNotFoundError(err)
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
	err = t.Client.Set(ctx, THROTTLERPREFIX+info.Email, encoded, throttlerService.THROTTLERSESSIONDURATION).Err()
	if err != nil {
		return rp.NewRessourceUpdateErr(err)
	}
	return nil
}
