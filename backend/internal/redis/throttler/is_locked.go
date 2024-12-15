package throttlerRepository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/throttler"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

// TODO: do the implementation for that function
func (t *Repository) IsLocked(ctx context.Context, key string) (bool, error) {
	var res throttlerService.Info
	val, err := t.Client.Get(ctx, THROTTLERPREFIX+key).Bytes()
	switch {
	case err == redis.Nil:
		return false, rp.NewNotFoundError(err)
	case err != nil:
		return false, rp.NewDatabaseErr(err)
	}

	json.Unmarshal(val, &res)
	res.Email = key

	return time.Now().Before(res.LockedUntil), nil
}
