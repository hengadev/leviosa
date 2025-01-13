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
            email,
            password,
            lastname,
            firstname,
            gender,
            birthdate,
            telephone,
            google_id,
            apple_id
        FROM users 
        WHERE email = ?;`

	err := u.DB.QueryRowContext(ctx, query, emailHash).Scan(
		&user.EmailHash,
		&user.PasswordHash,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.BirthDate,
		&user.Telephone,
		&user.GoogleID,
		&user.AppleID,
	)
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
