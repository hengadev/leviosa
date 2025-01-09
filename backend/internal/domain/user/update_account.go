package userService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/google/uuid"
)

func (s *Service) UpdateAccount(ctx context.Context, user *models.User, userID string) error {
	// validate UUID provided
	if err := uuid.Validate(userID); err != nil {
		return domain.NewInvalidValueErr(fmt.Sprintf("invalid user ID: %s", err.Error()))
	}
	// encrypt the user data here
	if errs := s.EncryptUser(user); len(errs) > 0 {
		return domain.NewInvalidValueErr(fmt.Sprintf("invalid user encryption: %s", errs.Error()))
	}
	// call modify account on the new data
	err := s.repo.ModifyAccount(
		ctx,
		user,
		map[string]any{"id": userID},
		prohibitedFields...,
	)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrInternal):
			fallthrough
		case errors.Is(err, rp.ErrContext):
			return err
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrNotUpdated):
			return domain.NewNotUpdatedErr(err)
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
