package userRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (u *Repository) GetPendingUser(ctx context.Context, emailHash string, provider models.ProviderType) (*models.User, error) {
	var user models.User
	var query string
	var args []interface{}
	switch provider {
	case models.Google:
	case models.Apple:
	case models.Mail:
		query = `
            SELECT 
                id,
                email,
                password,
                lastname,
                firstname,
                gender,
                birthdate,
                telephone,
                postal_code,
                city,
                address1,
                address2
            FROM pending_users 
            WHERE email = ?;`
		args = []interface{}{
			&user.ID,
			&user.EmailHash,
			&user.PasswordHash,
			&user.LastName,
			&user.FirstName,
			&user.Gender,
			&user.EncryptedBirthDate,
			&user.Telephone,
			&user.PostalCode,
			&user.City,
			&user.Address1,
			&user.Address2,
		}
	default:
		return nil, rp.NewValidationErr(fmt.Errorf("unsupported provider type: %v", provider), "provider")
	}

	err := u.DB.QueryRowContext(ctx, query, emailHash).Scan(args...)
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
