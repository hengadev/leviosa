package userRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) GetAllUsers(ctx context.Context) ([]*userService.User, error) {
	query := "SELECT email, role, lastname, firstname, gender, birthdate, telephone FROM users;"
	rows, err := u.DB.QueryContext(ctx, query)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextError(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	defer rows.Close()

	var users []*userService.User
	for rows.Next() {
		var user *userService.User
		err := rows.Scan(
			&user.Email,
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
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, rp.NewDatabaseErr(err)
	}
	if len(users) == 0 {
		return []*userService.User{}, nil
	}
	return users, nil
}
