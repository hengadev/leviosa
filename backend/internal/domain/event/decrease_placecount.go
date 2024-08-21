package event

import (
	"context"
	"fmt"
)

func (s *Service) DecreasePlacecount(ctx context.Context, eventID string) error {
	err := s.Repo.DecreaseFreeplace(ctx, eventID)
	if err != nil {
		return fmt.Errorf("decrease freeplace for event %s: %w", eventID, err)
	}
	return nil
}
