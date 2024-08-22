package userRepository

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *repository) GetAllUsers(ctx context.Context) ([]*userService.User, error) {
	var users []*userService.User
	query := "SELECT email, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard FROM users;"
	rows, err := u.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	defer rows.Close()

	for rows.Next() {
		user := &userService.User{}
		err := rows.Scan(
			&user.Email,
			&user.Role,
			&user.LastName,
			&user.FirstName,
			&user.Gender,
			&user.BirthDate,
			&user.Telephone,
			&user.Address,
			&user.City,
			&user.PostalCard,
		)
		if err != nil {
			return nil, rp.NewNotFoundError(err)
		}
		users = append(users, user)
	}
	return users, nil
}
