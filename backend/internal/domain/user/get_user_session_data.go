package userService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) GetUserSessionData(ctx context.Context, email string) (string, Role, error) {
	if _, pbms := NewEmail(email); len(pbms) > 0 {
		return "", UNKNOWN, domain.NewInvalidValueErr(fmt.Sprintf("invalid email: %q", pbms))
	}
	ID, role, err := s.repo.GetUserSessionData(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrNotFound):
			return "", UNKNOWN, domain.NewNotFoundErr(err)
		case errors.Is(err, rp.ErrContext):
			return "", UNKNOWN, err
		case errors.Is(err, rp.ErrDatabase):
			return "", UNKNOWN, domain.NewQueryFailedErr(err)
		default:
			return "", UNKNOWN, domain.NewUnexpectTypeErr(err)
		}
	}
	return ID, role, nil
}
