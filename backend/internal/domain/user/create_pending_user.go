package userService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/security"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// the OTP has been validated

func (s *Service) CreatePendingUser(ctx context.Context, email string) error {
	emailHash := security.HashEmail(email)
	user, err := s.repo.GetUnverifiedUser(ctx, emailHash)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrContext):
			return err
		case errors.Is(err, rp.ErrNotFound):
			return domain.NewNotFoundErr(err)
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewUnexpectTypeErr(err)
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	// add user to pending_user table
	if err = s.repo.AddPendingUser(ctx, user, models.Mail); err != nil {
		switch {
		case errors.Is(err, rp.ErrContext):
			return err
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrNotCreated):
			return domain.NewNotCreatedErr(err)
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
