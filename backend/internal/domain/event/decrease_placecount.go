package event

import (
	"context"
	"fmt"
)

func (s *Service) DecreasePlacecount(ctx context.Context, eventID string) error {
	if err := s.Repo.DecreaseFreeplace(ctx, eventID); err != nil {
		return fmt.Errorf("decrease freeplace for event: %w", err)
	}
	return nil
}
