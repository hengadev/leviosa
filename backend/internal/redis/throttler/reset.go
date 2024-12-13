package throttlerRepository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/throttler"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// Reset is called with a successful sign in to reset the throtlting associated with the provided email.
func (t *Repository) Reset(ctx context.Context, key string) error {
	var info throttlerService.Info
	val, err := t.Client.Get(ctx, THROTTLERPREFIX+key).Bytes()
	if err != nil {
		return rp.NewNotFoundError(err)
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
	err = t.Client.Set(ctx, THROTTLERPREFIX+info.Email, encoded, throttlerService.THROTTLERSESSIONDURATION).Err()
	if err != nil {
		return rp.NewRessourceUpdateErr(err)
	}
	return nil
}
