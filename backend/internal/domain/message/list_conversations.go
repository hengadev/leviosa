package messageService

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/message/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (s *Service) ListConversations(ctx context.Context, userID string) ([]*models.Conversation, error) {
	conversations, err := s.repo.ListConversations(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrNotCreated):
			return nil, domain.NewNotCreatedErr(err)
		case errors.Is(err, rp.ErrContext):
			return nil, err
		case errors.Is(err, rp.ErrDatabase):
			return nil, domain.NewQueryFailedErr(err)
		}
	}
	return conversations, nil
}
