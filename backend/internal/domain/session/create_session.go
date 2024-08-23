package sessionService

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

// that is basically the sign up of that function
func (s *Service) CreateSession(ctx context.Context, userID int, role string) (string, error) {
	session, err := NewSession(userID, role)
	if err != nil {
		return "", app.NewInvalidUserErr(err)
	}
	session.Create()
	session.Login()

	if err := s.Repo.CreateSession(ctx, session); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return session.ID, nil
}
