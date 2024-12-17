package userRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
)

func (u *Repository) AddAccount(ctx context.Context, usr *userService.User, provider ...string) (int64, error) {
	tx, err := u.DB.BeginTx(ctx, &sql.TxOptions{})
	// TODO: better handling for the rollback case
	defer func() {
		if err := tx.Rollback(); err != nil && errors.Is(err, sql.ErrTxDone) {
			// do some log thing if need be
		}
	}()
	if err != nil {
		return 0, fmt.Errorf("start transaction")
	}
	var nullString sql.NullString
	if usr.Password != "" {
		hashpassword, err := sqliteutil.HashPassword(usr.Password)
		if err != nil {
			return 0, rp.NewQueryErr(err)
		}
		nullString.Valid = true //NOTE: Do I need that ?j
		nullString.String = hashpassword
	}

	var providers string
	// if err := u.DB.QueryRowContext(ctx, "SELECT oauth_providers from users where email = ?;", usr.Email).Scan(providers); err != sql.ErrNoRows {
	if err := tx.QueryRowContext(ctx, "SELECT oauth_providers from users where email = ?;", usr.Email).Scan(providers); err != sql.ErrNoRows {
		return 0, rp.NewNotFoundError(err)
	}
	providerList := parseProviders(providers)
	addIfNotExist(providerList, usr.OAuthProvider)

	query := "INSERT INTO users (email, password, createdat, loggedinat, role, lastname, firstname, gender, birthdate, telephone, oauth_providers, oauth_ids) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"
	// res, err := u.DB.ExecContext(
	res, err := tx.ExecContext(
		ctx,
		query,
		usr.Email,
		nullString,
		usr.CreatedAt,
		usr.LoggedInAt,
		usr.Role,
		usr.LastName,
		usr.FirstName,
		usr.Gender,
		usr.BirthDate,
		usr.Telephone,
		formatProvider(providerList),
		usr.OAuthID,
	)
	if err != nil {
		return 0, rp.NewNotCreatedErr(err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	lastInsertID, err := res.LastInsertId()
	if lastInsertID == 0 {
		return 0, rp.NewNotCreatedErr(fmt.Errorf("no user added"))
	}
	return lastInsertID, nil
}

// function that given a list of string providers returns a formatted string of providers "providerA | providerB | providerC".
func formatProvider(providers []string) string {
	return strings.Join(providers, " | ")
}

// function that given a formatted list of providers "providerA | providerB | providerC" returns a list of providers
func parseProviders(providers string) []string {
	return strings.Split(providers, " | ")
}

func addIfNotExist(providers []string, provider string) []string {
	for _, p := range providers {
		if p == provider {
			return providers
		}
	}
	providers = append(providers, provider)
	return providers
}
