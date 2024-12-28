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
            created_at,
            postal_code,
            city,
            address1,
            address2
        ) VALUES (?, ?, ?, ?, ?, ?, ?, Datetime('now'),?,?,?,?);`
	result, err := u.DB.ExecContext(
		ctx,
		query,
		user.EmailHash,
		user.PasswordHash,
		user.LastName,
		user.FirstName,
		user.Gender,
		user.EncryptedBirthDate,
		user.Telephone,
		user.PostalCode,
		user.City,
		user.Address1,
		user.Address2,
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
