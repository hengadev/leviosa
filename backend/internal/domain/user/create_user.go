package userService

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CreateUser(ctx context.Context, userResponse *models.UserPendingResponse) (*models.User, error) {
	// NOTE: for now I just hard coder the provider but it should come from userResponse
	provider := models.Mail

	// get encrypted user from hashed email
	user, err := s.repo.GetPendingUser(ctx, userResponse.EmailHash)
	if err != nil {
		switch {
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

	// TODO: check if there is an existing oauth user in it
	// if there is a google_id just add the field that are not empty ?
	// same with apple_id

	// add user to database
	err = s.repo.AddUser(ctx, user, provider)
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
