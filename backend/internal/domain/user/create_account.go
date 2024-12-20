package userService

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

// TODO: Do the validation for the rest of the fields.
func (s *Service) CreateAccount(ctx context.Context, userCandidate *User) (*User, error) {
	var pbms errsx.Map
	email, emailPbms := NewEmail(userCandidate.Email)
	if len(emailPbms) > 0 {
		pbms.Set("email", emailPbms)
	}
	password, err := NewPassword(userCandidate.Password)
	if err != nil {
		pbms.Set("password", err)
	}
	telephone, err := NewTelephone(userCandidate.Telephone)
	if err != nil {
		pbms.Set("telephone", err)
	}

	if pbms != nil {
		return nil, pbms
	}

	user := NewUser(
		email,
		password,
		userCandidate.BirthDate,
		userCandidate.LastName,
		userCandidate.FirstName,
		userCandidate.Gender,
		telephone,
	)
	user.Create()
	user.Login()
	if err = s.repo.AddAccount(ctx, user); err != nil {
		return nil, domain.NewNotCreatedErr(err)
	}
	return user, nil
}
