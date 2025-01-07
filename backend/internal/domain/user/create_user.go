package userService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CreateUser(ctx context.Context, userResponse *models.UserPendingResponse) (*models.User, error) {
	// get encrypted user from hashed email
	user, err := s.repo.GetPendingUser(ctx, userResponse.Email, userResponse.Provider)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrValidation):
			return nil, domain.NewInvalidValueErr(err.Error())
		case errors.Is(err, rp.ErrNotFound):
			return nil, domain.NewNotFoundErr(err)
		case errors.Is(err, rp.ErrContext):
			return nil, err
		case errors.Is(err, rp.ErrDatabase):
			return nil, domain.NewQueryFailedErr(err)
		default:
			return nil, domain.NewUnexpectTypeErr(err)
		}
	}
	// convert to Role to see if there is no problem with the role sent
	role := models.ConvertToRole(userResponse.Role)
	if role == models.UNKNOWN {
		return nil, domain.NewInvalidValueErr("invalid role")
	}
	// add remaining field to user
	user.Role = role.String()
	user.Create()
	user.Login()

	// add user to database
	err = s.repo.AddUser(ctx, user, userResponse.Provider)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrNotCreated):
			return nil, domain.NewNotCreatedErr(err)
		case errors.Is(err, rp.ErrContext):
			return nil, err
		default:
			return nil, domain.NewUnexpectTypeErr(err)
		}
	}
	return user, nil
}
