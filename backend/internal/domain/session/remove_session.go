package sessionService

import (
	"context"
	"errors"
	"fmt"
)

func (s *Service) RemoveSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("nil sessionID")
	}
	if err := s.Repo.RemoveSession(ctx, sessionID); err != nil {
		return fmt.Errorf("repository error: %w", err)
	}
	return nil
}
