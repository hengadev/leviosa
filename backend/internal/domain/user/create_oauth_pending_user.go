package userService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (s *Service) CreateOAuthPendingUser(ctx context.Context, user *models.User, provider models.ProviderType) error {
	// encrypt user
	if errs := s.EncryptUser(user); len(errs) > 0 {
		return domain.NewInvalidValueErr(fmt.Sprintf("user encryption: %s", errs.Error()))
	}
	// add user to pending_user table
	if err := s.repo.AddPendingUser(ctx, user, provider); err != nil {
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
