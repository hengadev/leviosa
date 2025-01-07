package userRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) FindAccountByID(ctx context.Context, id string) (*models.User, error) {
	var user *models.User
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
	return user, nil
}
