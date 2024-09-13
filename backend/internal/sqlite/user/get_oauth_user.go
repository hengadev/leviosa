package userRepository

import (
	"context"
	"database/sql"
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
	if err == sql.ErrNoRows {
		return nil, rp.NewNotFoundError(err)
	}
	if err != nil {
		return nil, err
	}
	formattedProviders := parseProviders(providers)
	if !providerExists(formattedProviders, provider) {
		return nil, rp.NewNotFoundError(fmt.Errorf("provider %s not found in providers list", provider))
	}
	fmt.Println("the list of providers that I get is :", formattedProviders)
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
