package userRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

// GetPendingUsers retrieves all pending users from the database.
//
// Parameters:
//   - ctx: Context to manage the lifecycle of the operation and handle cancellation.
//
// Returns:
//   - []*models.User: A slice of pointers to user models populated with the retrieved data.
//   - error: An error if the query fails or no users are found.
//   - Returns a "not found" error if the result set is empty.
//   - Returns a context error if the operation is canceled or the deadline is exceeded.
//   - Returns a database error for any other query-related issues.
func (u *Repository) GetPendingUsers(ctx context.Context) ([]*models.User, error) {
	query := `
        SELECT 
            email,
            lastname,
            firstname,
            google_id,
            apple_id
        FROM users;`
	rows, err := u.DB.QueryContext(ctx, query)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.EmailHash,
			&user.LastName,
			&user.FirstName,
			&user.GoogleID,
			&user.AppleID,
		); err != nil {
			return nil, rp.NewDatabaseErr(err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, rp.NewDatabaseErr(err)
	}
	if len(users) == 0 {
		return []*models.User{}, rp.NewNotFoundErr(err, "pending users list")
	}
	return users, nil
}
