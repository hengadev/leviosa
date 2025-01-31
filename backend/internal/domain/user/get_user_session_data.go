package userService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/domain/user/security"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

// GetUserSessionData retrieves the session data for a user based on their email.
//
// Parameters:
//   - ctx: A context.Context instance to manage request lifecycle and cancellation.
//   - email: A string representing the email address of the user whose session data is being retrieved.
//
// Returns:
//   - string: The user's ID if found. If an error occurs or the user is not found, an empty string is returned.
//   - models.Role: The user's role. If an error occurs or the user is not found, models.UNKNOWN is returned.
//   - error: An error if the session data cannot be retrieved, the email is invalid, or an unexpected error occurs.
//     Returns nil if the session data is successfully retrieved.
func (s *Service) GetUserSessionData(ctx context.Context, email string) (string, models.Role, error) {
	if _, pbms := models.NewEmail(email); len(pbms) > 0 {
		return "", models.UNKNOWN, domain.NewInvalidValueErr(fmt.Sprintf("invalid email: %q", pbms))
	}
	hashedEmail := security.HashEmail(email)
	ID, role, err := s.repo.GetUserSessionData(ctx, hashedEmail)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrNotFound):
			return "", models.UNKNOWN, domain.NewNotFoundErr(err)
		case errors.Is(err, rp.ErrContext):
			return "", models.UNKNOWN, err
		case errors.Is(err, rp.ErrDatabase):
			return "", models.UNKNOWN, domain.NewQueryFailedErr(err)
		default:
			return "", models.UNKNOWN, domain.NewUnexpectTypeErr(err)
		}
	}
	return ID, role, nil
}
