package user

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

func (s *Service) CreateAccount(ctx context.Context, email, password string) (*User, error) {
	var input struct {
		Email    Email
		Password Password
	}
	{
		var err error
		// it is in the same package it does not make sense why that thing does not import
		if input.Email, err = NewEmail(email); err != nil {
			return nil, app.NewInvalidInputErr(err)
		}
		if input.Password, err = NewPassword(password); err != nil {
			return nil, app.NewInvalidInputErr(err)
		}
	}
	account := NewAccount(input.Email, input.Password)
	account.Create()
	account.Login()
	if err := s.repo.AddAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("add account: %w", err)
	}
	return account, nil
}
