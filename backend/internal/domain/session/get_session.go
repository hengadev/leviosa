package session

import (
	"context"
)

func (s *Service) GetSessionIDByUserID(ctx context.Context, userID int) (string, error) {
	sessionID, err := s.Repo.GetSessionIDByUserID(ctx, userID)
	if err != nil {
		return "", nil
	}
	return sessionID, nil
}
