package userRepository

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// reader
func (u *UserRepository) FindAccountByID(ctx context.Context, id int) (*user.User, error) {
	var user user.User
	if err := u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?;", id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
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
	); err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	return &user, nil
}
