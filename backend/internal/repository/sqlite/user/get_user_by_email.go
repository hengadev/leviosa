package userRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

// GetUserByEmail retrieves a user's details from the database using their email hash.
//
// Parameters:
//   - ctx: Context for managing the lifecycle of the operation (e.g., handling timeouts and cancellations).
//   - emailHash: The hashed email of the user to search for.
//
// Returns:
//   - *models.User: A pointer to a populated User model if the user is found.
//   - error: An error if the query fails or no matching user is found.
//   - Returns a "not found" error if no user matches the provided email hash.
//   - Returns a context error if the operation is canceled or times out.
//   - Returns a database error for other query-related issues.
func (u *Repository) GetUserByEmail(ctx context.Context, emailHash string) (*models.User, error) {
	var user models.User
	query := `
        SELECT 
            encrypted_email,
            encrypted_picture,
            encrypted_created_at,
            encrypted_logged_in_at,
            encrypted_role,
            encrypted_birthdate,
            encrypted_lastname,
            encrypted_firstname,
            encrypted_gender,
            encrypted_telephone,
            encrypted_postal_code,
            encrypted_city,
            encrypted_address1,
            encrypted_address2,
            encrypted_google_id,
            encrypted_apple_id
        FROM users 
        WHERE email_hash = ?;`

	err := u.DB.QueryRowContext(ctx, query, emailHash).Scan(
		&user.Email,
		&user.Picture,
		&user.EncryptedCreatedAt,
		&user.EncryptedLoggedInAt,
		&user.Role,
		&user.EncryptedBirthDate,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.Telephone,
		&user.PostalCode,
		&user.City,
		&user.Address1,
		&user.Address2,
		&user.GoogleID,
		&user.AppleID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, rp.NewNotFoundErr(err, "user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return &user, nil
}
