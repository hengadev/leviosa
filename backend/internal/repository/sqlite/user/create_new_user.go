package userRepository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

// createNewUser inserts a new user into the specified table ('pending_users', 'users', etc.)
// based on the authentication provider type. The function prepares the appropriate SQL query
// for inserting user data based on the provider (Google, Apple, or Mail) and executes it within
// a transaction if provided.
//
// Parameters:
//   - ctx: The context for managing the transaction lifecycle and cancelation.
//   - tx: The transaction object used for the operation. If nil, the operation is executed outside of a transaction.
//   - user: The user object containing details such as email hash, password hash, personal details, and provider-specific information.
//   - provider: The authentication provider type (Google, Apple, or Mail) which determines how the user data is inserted into the table.
//   - table: The name of the table into which the user data is inserted (e.g., 'users', 'pending_users').
//
// Returns:
//   - error: An error if the user insertion fails, including database-related errors.
//   - If the provider type is unsupported, a validation error is returned.
//   - If the insert operation does not affect any rows, an error indicating no creation is returned.
func (u *Repository) createNewUser(ctx context.Context, tx *sql.Tx, user *models.User, provider models.ProviderType, table string) error {
	var query string
	var args []interface{}

	switch provider {
	// TODO: add the remaining fields for address that are missing from google here
	case models.Google:
		query = fmt.Sprintf(`
            INSERT INTO %s (
                id,
                email_hash,
                encrypted_email,
                password_hash,
                encrypted_lastname,
                encrypted_firstname,
                encrypted_gender,
                encrypted_birthdate,
                encrypted_telephone,
                encrypted_google_id,
                encrypted_apple_id
            ) VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, NULL)`, table)
		args = []interface{}{
			user.ID,
			user.EmailHash,
			user.Email,
			user.LastName,
			user.FirstName,
			user.Gender,
			user.EncryptedBirthDate,
			user.Telephone,
			user.GoogleID,
		}
	case models.Apple:
		query = fmt.Sprintf(`
            INSERT INTO %s (
                id,
                email_hash,
                encrypted_email,
                password_hash,
                encrypted_lastname,
                encrypted_firstname,
                encrypted_gender,
                encrypted_birthdate,
                encrypted_telephone,
                encrypted_google_id,
                encrypted_apple_id
            ) VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, NULL, ?)`, table)
		args = []interface{}{
			user.ID,
			user.EmailHash,
			user.Email,
			user.LastName,
			user.FirstName,
			user.Gender,
			user.EncryptedBirthDate,
			user.Telephone,
			user.AppleID,
		}
	case models.Mail:
		query = fmt.Sprintf(`
            INSERT INTO %s (
                id,
                email_hash,
                encrypted_email,
                password_hash,
                encrypted_lastname,
                encrypted_firstname,
                encrypted_gender,
                encrypted_birthdate,
                encrypted_telephone,
                encrypted_postal_code,
                encrypted_city,
                encrypted_address1,
                encrypted_address2,
                encrypted_google_id,
                encrypted_apple_id
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL)`, table)
		args = []interface{}{
			user.ID,
			user.EmailHash,
			user.Email,
			user.Password,
			user.LastName,
			user.FirstName,
			user.Gender,
			user.EncryptedBirthDate,
			user.Telephone,
			user.PostalCode,
			user.City,
			user.Address1,
			user.Address2,
		}
	default:
		return rp.NewValidationErr(fmt.Errorf("unsupported provider type: %v", provider), "provider")
	}
	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, args...)
	} else {
		result, err = u.DB.ExecContext(ctx, query, args...)
	}
	if err != nil {
		return rp.NewNotCreatedErr(fmt.Errorf("failed to create user in %s table: %w", table, err), "pending user")
	}

	// Check if the insert was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(err, fmt.Sprintf("new user in %s table", table))
	}

	return nil
}
