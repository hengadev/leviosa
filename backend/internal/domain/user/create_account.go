package userService

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

// TODO: Do the validation for the rest of the fields.
func (s *Service) CreateAccount(ctx context.Context, userCandidate *User) (*User, error) {
	var input struct {
		Email     Email
		Password  Password
		Telephone Telephone
	}
	{
		var err error
		if input.Email, err = NewEmail(userCandidate.Email); err != nil {
			return nil, app.NewInvalidInputErr(err)
		}
		if input.Password, err = NewPassword(userCandidate.Password); err != nil {
			return nil, app.NewInvalidInputErr(err)
		}
		if input.Telephone, err = NewTelephone(userCandidate.Telephone); err != nil {
			return nil, app.NewInvalidInputErr(err)
		}
	}
	user := NewUser(
		input.Email,
		input.Password,
		userCandidate.BirthDate,
		userCandidate.LastName,
		userCandidate.FirstName,
		userCandidate.Gender,
		input.Telephone,
	)
	user.Create()
	user.Login()
	id, err := s.repo.AddAccount(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("add account: %w", err)
	}
	user.ID = int(id)
	return user, nil
}
