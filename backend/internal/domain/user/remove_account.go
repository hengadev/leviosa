package userService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/google/uuid"
)

func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	if err := uuid.Validate(userID); err != nil {
		return domain.NewInvalidValueErr(fmt.Sprintf("invalid user ID: %s", err))
	}
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		switch {
		case errors.Is(err, rp.ErrContext):
			return err
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrNotDeleted):
			return domain.NewNotDeletedErr(err)
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}

	return nil
}
