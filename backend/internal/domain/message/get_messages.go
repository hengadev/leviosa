package messageService

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/message/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (s *Service) GetMessages(ctx context.Context, conversationID string) ([]*models.Message, error) {
	messages, err := s.repo.GetMessages(ctx, conversationID)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrNotFound):
			return nil, domain.NewNotFoundErr(err)
		case errors.Is(err, rp.ErrContext):
			return nil, err
		case errors.Is(err, rp.ErrDatabase):
			return nil, domain.NewQueryFailedErr(err)
		}
	}
	for _, message := range messages {
		if errs := s.DecryptMessage(message); len(errs) > 0 {
			return nil, domain.NewNotEncryptedErr("message content decryption", errs)
		}
	}
	return messages, nil
}
