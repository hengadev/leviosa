package userService

import (
	"context"
	"errors"

	app "github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) GetUserSessionData(ctx context.Context, email string) (string, Role, error) {
	ID, role, err := s.repo.GetUserSessionData(ctx, email)
	switch {
	case errors.Is(err, rp.ErrNotFound):
		return "", UNKNOWN, app.NewUserNotFoundErr(err)
	case errors.Is(err, rp.ErrDatabase):
		return "", UNKNOWN, app.NewQueryFailedErr(err)
	case err != nil:
		return "", UNKNOWN, app.NewUnexpectTypeErr(err)
	}

	return ID, role, nil
}
