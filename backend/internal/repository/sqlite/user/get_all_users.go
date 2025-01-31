package userRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

// GetAllUsers retrieves all user accounts from the 'users' table.
// The function queries the database to fetch the relevant details for each user
// and maps the results to a slice of User models.
//
// Parameters:
//   - ctx: The context for managing the transaction lifecycle and cancelation.
//
// Returns:
//   - []*models.User: A slice of pointers to User models containing the retrieved user details.
//   - error: An error if the query fails or other issues occur during the operation.
//   - If the operation is canceled or the deadline is exceeded, a context error is returned.
//   - For query failures or result processing issues, a database error is returned.
//   - Returns an empty slice with no error if no users are found.
func (u *Repository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	query := "SELECT email, role, lastname, firstname, gender, birthdate, telephone FROM users;"
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
		err := rows.Scan(
			&user.EmailHash,
			&user.Role,
			&user.LastName,
			&user.FirstName,
			&user.Gender,
			&user.BirthDate,
			&user.Telephone,
		)
		if err != nil {
			return nil, rp.NewDatabaseErr(err)
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, rp.NewDatabaseErr(err)
	}
	if len(users) == 0 {
		return []*models.User{}, nil
	}
	return users, nil
}
