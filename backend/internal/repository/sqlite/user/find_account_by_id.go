package userRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

// FindAccountByID retrieves a user's account details from the 'users' table based on the provided user ID.
// The function performs a database query to fetch the user's data and maps it to a User model.
// If the user is not found or an error occurs during the operation, appropriate errors are returned.
//
// Parameters:
//   - ctx: The context for managing the transaction lifecycle and cancelation.
//   - id: The unique identifier of the user to find in the database.
//
// Returns:
//   - *models.User: A pointer to the User model containing the retrieved account details.
//   - error: An error if the query fails or the user is not found.
//   - If the query returns no rows, a "not found" error is returned.
//   - If context-related errors occur (e.g., deadline exceeded, canceled), a context error is returned.
//   - For any other query failures, a database error is returned.
func (u *Repository) FindAccountByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	query := `
        SELECT 
            email,
            picture,
            role,
            lastname,
            firstname,
            gender,
            birthdate,
            telephone,
            postal_code,
            city,
            address1,
            address2
        FROM users
        WHERE id = ?;`
	if err := u.DB.QueryRowContext(ctx, query, id).Scan(
		&user.EmailHash,
		&user.Picture,
		&user.Role,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.BirthDate,
		&user.Telephone,
		&user.PostalCode,
		&user.City,
		&user.Address1,
		&user.Address2,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return &models.User{}, rp.NewNotFoundErr(err, "unverified user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return &user, nil
}
