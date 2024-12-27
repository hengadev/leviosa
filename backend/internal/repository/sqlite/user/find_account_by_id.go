package userRepository

import (
	"context"
	"database/sql"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) FindAccountByID(ctx context.Context, id string) (*models.User, error) {
	var nullPassword sql.NullString
	var user models.User
	if err := u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?;", id).Scan(
		&user.ID,
		&user.EmailHash,
		&nullPassword,
		&user.CreatedAt,
		&user.LoggedInAt,
		&user.Role,
		&user.BirthDate,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.Telephone,
		&user.GoogleID,
		&user.AppleID,
	); err != nil {
		return nil, rp.NewNotFoundErr(err, "user by ID")
	}
	// get the passowrd in the user instance if not null
	if nullPassword.Valid {
		user.PasswordHash = nullPassword.String
	}
	return &user, nil
}
