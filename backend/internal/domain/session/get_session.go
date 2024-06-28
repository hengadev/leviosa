package session

import (
	"context"
)

func (s *Service) GetSessionIDByUserID(ctx context.Context, userID string) (string, error) {
	sessionID, err := s.repo.GetSessionIDByUserID(ctx, userID)
	if err != nil {
		return "", nil
	}
	return sessionID, nil
}
