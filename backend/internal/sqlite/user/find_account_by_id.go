package userRepository

import (
	"context"
	"database/sql"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) FindAccountByID(ctx context.Context, id int) (*userService.User, error) {
	var nullString sql.NullString
	var user userService.User
	if err := u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?;", id).Scan(
		&user.ID,
		&user.Email,
		&nullString,
		&user.CreatedAt,
		&user.LoggedInAt,
		&user.Role,
		&user.BirthDate,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.Telephone,
		&user.Address,
		&user.City,
		&user.PostalCard,
		&user.OAuthProvider,
		&user.OAuthID,
	); err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	// get the passowrd in the user instance if not null
	if nullString.Valid {
		user.Password = nullString.String
	}
	return &user, nil
}
