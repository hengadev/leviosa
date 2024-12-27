package userService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/security"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// TODO: change that thing using the new API
func (s *Service) ValidateCredentials(ctx context.Context, user *models.UserSignIn) error {
	hashedEmail := security.HashEmail(user.Email)
	hashedPassword, err := s.repo.GetHashedPasswordByEmail(ctx, hashedEmail)
	if err != nil {
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
	ok, err := s.VerifyPassword(user.Password, hashedPassword)
	if err != nil {
		// error in the verification
	}
	if !ok {
		// the password are not the same
	}
	return nil
}
