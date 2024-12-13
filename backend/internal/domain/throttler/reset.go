package throttlerService

import (
	"context"
	"fmt"
)

func (s *Service) Reset(ctx context.Context, email string) error {
	// TODO: make better error handling for this function
	isLocked, err := s.repo.IsLocked(ctx, email)
	if err != nil {
		return fmt.Errorf("locked for the current email")
	}
	if isLocked {
		return fmt.Errorf("user locked, too many request: %w", err)
	}
	if err := s.repo.Reset(ctx, email); err != nil {
		return fmt.Errorf("reset attempt: %w", err)
	}
	return nil
}
