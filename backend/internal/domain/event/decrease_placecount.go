package event

import (
	"context"
	"fmt"
)

func (s *Service) DecreasePlacecount(ctx context.Context, eventID string) error {
	rowsAffected, err := s.Repo.DecreaseFreeplace(ctx, eventID)
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("decrease freeplace for event %s: %w", eventID, err)
	}
	return nil
}
