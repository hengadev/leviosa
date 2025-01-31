package userService

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/domain/user/security"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/google/uuid"
)

// CreatePendingUser creates a pending user entry in the system for email-based registration.
//
// Parameters:
//   - ctx: A context.Context instance to manage request lifecycle and cancellation.
//   - email: A string representing the email address of the user to be added to the pending user table.
//
// Returns:
//   - error: An error if the user cannot be retrieved, the user cannot be added to the pending user table,
//     or an unexpected error occurs. Returns nil if the user is successfully added.
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
			return domain.NewQueryFailedErr(err)
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}

	// creating the userID here
	user.ID = uuid.NewString()

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
