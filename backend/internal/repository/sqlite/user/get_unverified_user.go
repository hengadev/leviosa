package userRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hengadev/leviosa/internal/domain/user/models"
	rp "github.com/hengadev/leviosa/internal/repository"
)

// GetUnverifiedUser retrieves an unverified user's details by their email hash from the database.
//
// Parameters:
//   - ctx: Context to manage the lifecycle of the operation and handle cancellation.
//   - emailHash: The hashed email of the user to search for.
//
// Returns:
//   - *models.User: A pointer to the populated user model if the user is found.
//   - error: An error if the query fails or no matching user is found.
//   - Returns a "not found" error if no user matches the provided email hash.
//   - Returns a context error if the operation is canceled or the deadline is exceeded.
//   - Returns a database error for any other query-related issues.
func (u *Repository) GetUnverifiedUser(ctx context.Context, emailHash string) (*models.User, error) {
	var user models.User
	query := `
        SELECT 
            encrypted_email,
            encrypted_lastname,
            encrypted_firstname,
            encrypted_gender,
            encrypted_birthdate,
            encrypted_telephone,
            encrypted_postal_code,
            encrypted_city,
            encrypted_address1,
            encrypted_address2
        FROM unverified_users 
        WHERE email_hash = ?;`

	err := u.DB.QueryRowContext(ctx, query, emailHash).Scan(
		&user.Email,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.EncryptedBirthDate,
		&user.Telephone,
		&user.PostalCode,
		&user.City,
		&user.Address1,
		&user.Address2,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, rp.NewNotFoundErr(err, "unverified user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return &user, nil
}
