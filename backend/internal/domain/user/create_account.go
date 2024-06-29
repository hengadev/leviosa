package user

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

// TODO: I need mor field to create the account for the user
// func (s *Service) CreateAccount(ctx context.Context, email, password string) (*User, error) {
func (s *Service) CreateAccount(ctx context.Context, userCandidate *User) (*User, error) {
	var input struct {
		Email    Email
		Password Password
	}
	{
		var err error
		// if input.Email, err = NewEmail(email); err != nil {
		if input.Email, err = NewEmail(userCandidate.Email); err != nil {
			return nil, app.NewInvalidInputErr(err)
		}
		// if input.Password, err = NewPassword(password); err != nil {
		if input.Password, err = NewPassword(userCandidate.Password); err != nil {
			return nil, app.NewInvalidInputErr(err)
		}
	}
	account := NewUser(
		input.Email,
		input.Password,
		userCandidate.BirthDate,
		userCandidate.LastName,
		userCandidate.FirstName,
		userCandidate.Gender,
		userCandidate.Telephone,
		userCandidate.Address,
		userCandidate.City,
		userCandidate.PostalCard,
	)
	account.Create()
	account.Login()
	if err := s.repo.AddAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("add account: %w", err)
	}
	return account, nil
}
