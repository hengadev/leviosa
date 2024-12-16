package userService

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

// TODO: Do the validation for the rest of the fields.
func (s *Service) CreateAccount(ctx context.Context, userCandidate *User) (*User, error) {
	var pbms errsx.Map
	email, err := NewEmail(userCandidate.Email)
	if err != nil {
		pbms.Set("email", err)
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
	id, err := s.repo.AddAccount(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("add account: %w", err)
	}
	user.ID = string(id)
	return user, nil
}
