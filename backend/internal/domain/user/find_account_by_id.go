package userService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// FindAccountByID return the decrypted user with the specified ID
func (s *Service) FindAccountByID(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.repo.FindAccountByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrNotFound):
			return &models.User{}, nil
		case errors.Is(err, rp.ErrContext):
			return nil, err
		case errors.Is(err, rp.ErrDatabase):
			return nil, domain.NewQueryFailedErr(err)
		default:
			return nil, domain.NewUnexpectTypeErr(err)
		}
	}
	if errs := s.DecryptUser(user); len(errs) > 0 {
		return nil, domain.NewInvalidValueErr(err.Error())
	}
	return user, nil
}
