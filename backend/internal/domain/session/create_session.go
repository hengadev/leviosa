package sessionService

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

func (s *Service) CreateSession(ctx context.Context, userID string, role userService.Role) (*Session, error) {
	session, err := NewSession(userID, role)
	if err != nil {
		return nil, fmt.Errorf("create session object: %w", err)
	}
	if err := s.Repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("create session in redis: %w", err)
	}
	return session, nil
}
