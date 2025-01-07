package userService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) GetAllPendingUsers(ctx context.Context) ([]*models.UserPending, error) {
	users, err := s.repo.GetPendingUsers(ctx)
	var pendingUsers []*models.UserPending
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrContext):
			return nil, err
		case errors.Is(err, rp.ErrDatabase):
			return nil, domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrNotFound):
			return pendingUsers, nil
		default:
			return nil, domain.NewUnexpectTypeErr(err)
		}
	}
	for _, user := range users {
		if errs := s.DecryptUser(user); len(errs) > 0 {
			return nil, domain.NewInvalidValueErr(err.Error())
		}
		pendingUsers = append(pendingUsers, user.ToUserPending())
	}
	return pendingUsers, nil
}
