package userRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (u *Repository) GetUserByEmail(ctx context.Context, emailHash string) (*models.User, error) {
	var user models.User
	query := `
        SELECT 
            email,
            password,
            lastname,
            firstname,
            gender,
            birthdate,
            telephone,
            google_id,
            apple_id
        FROM users 
        WHERE email = ?;`

	err := u.DB.QueryRowContext(ctx, query, emailHash).Scan(
		&user.EmailHash,
		&user.PasswordHash,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.BirthDate,
		&user.Telephone,
		&user.GoogleID,
		&user.AppleID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, rp.NewNotFoundErr(err, "pending user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	return &user, nil
}
