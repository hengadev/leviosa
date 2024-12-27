package userService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/security"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CheckUser(ctx context.Context, email string) error {
	// hash email
	emailHash := security.HashEmail(email)
	// look for email in database
	if err := s.repo.HasUser(ctx, emailHash); err != nil {
		switch {
		case errors.Is(err, rp.ErrNotFound):
			return domain.NewNotFoundErr(err)
		case errors.Is(err, rp.ErrContext):
			return err
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
