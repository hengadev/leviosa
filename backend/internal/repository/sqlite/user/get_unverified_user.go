package userRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) GetUnverifiedUser(ctx context.Context, emailHash string) (*models.User, error) {
	var user models.User
	query := `
        SELECT 
            email,
            password,
            lastname,
            firstname,
            gender,
            birthdate,
            telephone
        FROM unverified_users 
        WHERE email = ?;`

	err := u.DB.QueryRowContext(ctx, query, emailHash).Scan(
		&user.EmailHash,
		&user.PasswordHash,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.BirthDate,
		&user.Telephone,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, rp.NewNotFoundErr(err, "unverified user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return &user, nil
}
