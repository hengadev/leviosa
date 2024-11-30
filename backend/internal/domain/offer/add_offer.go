package offerService

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) AddOffer(
	ctx context.Context,
	name string,
	description string,
	duration time.Time,
) error {
	offer := NewOffer(name, description, duration)
	// TODO: valid the offer thing or put inside the New function
	if err := s.repo.CreateOffer(ctx, offer); err != nil {
		return fmt.Errorf("create offfer")
	}
	return nil
}
