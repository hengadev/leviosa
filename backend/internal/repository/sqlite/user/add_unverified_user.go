package userRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

// AddUnverifiedUser inserts a new unverified user into the 'unverified_users' table.
//
// Parameters:
//   - ctx: The context for managing the transaction lifecycle and cancelation.
//   - user: The user object containing details to be stored in the 'unverified_users' table.
//     This includes email hash, password hash, personal details, and encrypted birthdate.
//
// Returns:
//   - error: An error if the insertion fails, including database or context-related errors. Returns nil if successful.
//   - If no rows are affected by the insertion, a "not created" error is returned.
func (u *Repository) AddUnverifiedUser(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO unverified_users (
            email_hash,
            encrypted_email,
            password_hash,
            encrypted_lastname,
            encrypted_firstname,
            encrypted_gender,
            encrypted_birthdate,
            encrypted_telephone,
            encrypted_created_at,
            encrypted_postal_code,
            encrypted_city,
            encrypted_address1,
            encrypted_address2
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	result, err := u.DB.ExecContext(
		ctx,
		query,
		user.EmailHash,
		user.Email,
		user.PasswordHash,
		user.LastName,
		user.FirstName,
		user.Gender,
		user.EncryptedBirthDate,
		user.Telephone,
		user.PostalCode,
		user.City,
		user.Address1,
		user.Address2,
	)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}

	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "unverified user with provided emailHash")
	}
	return nil
}
