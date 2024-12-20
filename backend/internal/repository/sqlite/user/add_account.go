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

func (u *Repository) AddAccount(ctx context.Context, usr *userService.User, provider ...string) error {
	tx, err := u.DB.BeginTx(ctx, &sql.TxOptions{})
	// TODO: better handling for the rollback case, using some panic recovery
	defer func() {
		if err := tx.Rollback(); err != nil && errors.Is(err, sql.ErrTxDone) {
			// do some log thing if need be
		}
	}()
	if err != nil {
		return rp.NewDatabaseErr(fmt.Errorf("start transaction: %w", err))
	}
	var nullString sql.NullString
	if usr.Password != "" {
		hashpassword, err := sqliteutil.HashPassword(usr.Password)
		if err != nil {
			return rp.NewInternalError(err)
		}
		nullString.Valid = true
		nullString.String = hashpassword
	}

	var providers string
	err = tx.QueryRowContext(ctx, "SELECT oauth_providers from users where email = ?;", usr.Email).Scan(providers)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return rp.NewNotFoundError(err, "user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	providerList := parseProviders(providers)
	addIfNotExist(providerList, usr.OAuthProvider)

	query := "INSERT INTO users (email, password, createdat, loggedinat, role, lastname, firstname, gender, birthdate, telephone, oauth_providers, oauth_ids) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"
	result, err := tx.ExecContext(
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
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextError(err)
		default:
			return rp.NewDatabaseErr(err)
		}

	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "user")
	}

	if err := tx.Commit(); err != nil {
		return rp.NewDatabaseErr(fmt.Errorf("commit transaction: %w", err))
	}
	return nil
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
