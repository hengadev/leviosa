package sessionService

import (
	"context"
	"errors"
	"fmt"
)

func (s *Service) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, errors.New("nil sessionID")
	}
	session, err := s.Repo.FindSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}
	return session, nil
}
