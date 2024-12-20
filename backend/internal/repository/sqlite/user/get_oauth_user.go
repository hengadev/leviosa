package userRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) GetOAuthUser(ctx context.Context, email, provider string) (*userService.User, error) {
	var user userService.User
	var providers, ids string
	err := u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE email = ?", email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.LoggedInAt,
		&user.Role,
		&user.BirthDate,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.Telephone,
		&providers,
		&ids,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, rp.NewNotFoundError(err, "oauth user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextError(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	formattedProviders := parseProviders(providers)
	if !providerExists(formattedProviders, provider) {
		return nil, rp.NewNotFoundError(fmt.Errorf("provider %s not found in providers list", provider), "provider")
	}

	return &user, nil
}

func providerExists(providers []string, providerName string) bool {
	for _, provider := range providers {
		if provider == providerName {
			return true
		}
	}
	return false
}
