package userRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

// GetPendingUser retrieves a pending user by their email hash and provider type from the 'pending_users' table.
//
// Parameters:
//   - ctx: The context for managing the transaction lifecycle and cancelation.
//   - emailHash: The hashed email address of the user to be retrieved.
//   - provider: The type of provider (e.g., Google, Apple, Mail) associated with the user.
//
// Returns:
//   - *models.User: A pointer to the user model populated with the retrieved data.
//   - error: An error if the query fails or the user is not found.
//   - Returns a "validation" error if the provider type is unsupported.
//   - Returns a "not found" error if no user exists with the given email hash.
//   - Returns a context error if the operation is canceled or the deadline is exceeded.
//   - Returns a database error for any other query-related issues.
func (u *Repository) GetPendingUser(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
	var user models.User
	var query string
	var args []interface{}
	switch provider {
	case models.Google:
	case models.Apple:
	case models.Mail:
		query = `
            SELECT 
                id,
                email,
                password,
                lastname,
                firstname,
                gender,
                birthdate,
                telephone,
                postal_code,
                city,
                address1,
                address2
            FROM pending_users 
            WHERE email = ?;`
		args = []interface{}{
			&user.ID,
			&user.EmailHash,
			&user.PasswordHash,
			&user.LastName,
			&user.FirstName,
			&user.Gender,
			&user.EncryptedBirthDate,
			&user.Telephone,
			&user.PostalCode,
			&user.City,
			&user.Address1,
			&user.Address2,
		}
	default:
		return nil, rp.NewValidationErr(fmt.Errorf("unsupported provider type: %v", provider), "provider")
	}

	err := u.DB.QueryRowContext(ctx, query, emailHash).Scan(args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, rp.NewNotFoundErr(err, "pending user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return &user, nil
}
