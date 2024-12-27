package userRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) AddUnverifiedUser(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO unverified_users (
            email,
            password,
            lastname,
            firstname,
            gender,
            birthdate,
            telephone,
            created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, NOW());`
	result, err := u.DB.ExecContext(
		ctx,
		query,
		user.Email,
		user.Password,
		user.LastName,
		user.FirstName,
		user.Gender,
		user.BirthDate,
		user.Telephone,
	)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}

	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "unverified user with provided emailHash")
	}
	return nil
}

