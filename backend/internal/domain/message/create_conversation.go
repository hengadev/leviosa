package messageService

import (
	"context"
	"errors"
	"time"

	"github.com/hengadev/leviosa/internal/domain"
	"github.com/hengadev/leviosa/internal/domain/message/models"
	rp "github.com/hengadev/leviosa/internal/repository"
)

func (s *Service) CreateConversation(ctx context.Context, userID, adminID string) (string, error) {
	// TODO: check if the userID sent has the right priviledge to be talked to, ie is freelance ?
	conversation := &models.Conversation{
		UserID:    userID,
		AdminID:   adminID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateConversation(ctx, conversation); err != nil {
		switch {
		case errors.Is(err, rp.ErrNotCreated):
			return "", domain.NewNotCreatedErr(err)
		case errors.Is(err, rp.ErrContext):
			return "", err
		case errors.Is(err, rp.ErrDatabase):
			return "", domain.NewQueryFailedErr(err)
		}
	}

	return conversation.ID, nil
}
