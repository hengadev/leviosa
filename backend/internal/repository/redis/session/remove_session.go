package sessionRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Repository) RemoveSession(ctx context.Context, ID string) error {
	res, err := s.client.Exists(ctx, SESSIONPREFIX+ID).Result()
	if err != nil {
		return rp.NewNotFoundError(err)
	}
	if res == 0 {
		return fmt.Errorf("non existing key")
	}
	if err := s.client.Del(ctx, SESSIONPREFIX+ID).Err(); err != nil {
		return rp.NewRessourceDeleteErr(err)
	}
	return nil
}
