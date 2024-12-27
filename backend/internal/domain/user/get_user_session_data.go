package userService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/security"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

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
