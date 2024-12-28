package userRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// AddGenericUser adds a user to the specified table taking into account the provider type specified.
// The function assure account linking if other providers exists.
func (u *Repository) addGenericUser(ctx context.Context, user *models.User, provider models.ProviderType, table string) error {
	tx, err := u.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return rp.NewDatabaseErr(fmt.Errorf("failed to start transaction: %w", err))
	}
	defer tx.Rollback()

	if !provider.IsValid() {
		return rp.NewValidationErr(fmt.Errorf("provider type value can only be 'apple', 'google' or 'mail', got : %q", provider), "provider type")
	}

	// Check if user exists
	var userID string
	err = tx.QueryRowContext(ctx,
		fmt.Sprintf("SELECT id FROM %s WHERE email = ?", table),
		user.EmailHash,
	).Scan(&userID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// Create new user
		err = u.createNewUser(ctx, tx, user, provider, table)
		if err != nil {
			return rp.NewNotCreatedErr(err, "new user")
		}
	} else {
		// Link authentication method to existing user
		if err := u.linkAuthMethod(ctx, tx, userID, user, provider, table); err != nil {
			return rp.NewNotCreatedErr(err, "link auth methods")
		}
	}

	if err := tx.Commit(); err != nil {
		return rp.NewDatabaseErr(fmt.Errorf("commit transaction: %w", err))
	}
	return nil
}
