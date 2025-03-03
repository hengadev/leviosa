package userService

import (
	"context"
	"errors"

	"github.com/hengadev/leviosa/internal/domain"
	"github.com/hengadev/leviosa/internal/domain/user/models"
	"github.com/hengadev/leviosa/internal/domain/user/security"
	rp "github.com/hengadev/leviosa/internal/repository"
)

// CheckUser verifies if a user with the given email exists in the system.
//
// Parameters:
//   - ctx: A context.Context instance to manage request lifecycle and cancellation.
//   - email: A string representing the email address of the user to check.
//
// Returns:
//   - error: An error if the user does not exist, the query fails, or an unexpected error occurs.
//     Returns nil if the user exists.
func (s *Service) CheckUser(ctx context.Context, email string) error {
	if errs := models.ValidateEmail(email); len(errs) > 0 {
		return domain.NewInvalidValueErr(errs.Error())
	}
	emailHash := security.HashEmail(email)
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
