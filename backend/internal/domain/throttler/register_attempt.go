package throttlerService

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) RegisterAttempt(ctx context.Context, email string) error {
	// TODO: make better error handling for this function
	now := time.Now()
	isLocked, err := s.repo.IsLocked(ctx, email)
	if err != nil {
		return fmt.Errorf("locked for the current email")
	}
	if isLocked {
		return fmt.Errorf("user locked, too many request: %w", err)
	}
	if err := s.repo.MakeAttempt(ctx, email, now); err != nil {
		return fmt.Errorf("user made a login attempt: %w", err)
	}
	return nil
}
