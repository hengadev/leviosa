package session

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

// that is basically the sign up of that function
func (s *Service) CreateSession(ctx context.Context, userID, role string) (string, error) {
	sessionID, err := s.repo.GetSessionIDByUserID(ctx, userID)
	if err == nil && sessionID != "" {
		return sessionID, nil
	}

	session, err := NewSession(userID, role)
	if err != nil {
		return "", app.NewInvalidUserErr(err)
	}

	sessionID, err = s.repo.CreateSession(ctx, session)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return sessionID, nil
}
