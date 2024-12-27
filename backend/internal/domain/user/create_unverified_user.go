package userService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CreateUnverifiedUser(ctx context.Context, userSignUp *models.UserSignUp) (string, error) {
	user := userSignUp.ToUser()
	if errs := s.EncryptUser(user); len(errs) > 0 {
		return "", domain.NewNotEncryptedErr(errs)
	}
	err := s.repo.AddUnverifiedUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrDatabase):
			return "", domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrNotCreated):
			return "", domain.NewNotCreatedErr(err)
		case errors.Is(err, rp.ErrContext):
			return "", err
		default:
			return "", domain.NewUnexpectTypeErr(err)
		}
	}
	return user.EmailHash, nil
}
