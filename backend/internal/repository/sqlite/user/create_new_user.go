package userRepository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (u *Repository) createNewUser(ctx context.Context, tx *sql.Tx, user *models.User, provider models.ProviderType, table string) error {
	var query string
	var args []interface{}

	switch provider {
	// TODO: add the remaining fields for address that are missing from google here
	case models.Google:
		query = fmt.Sprintf(`
            INSERT INTO %s (
                id,
                email,
                password,
                lastname,
                firstname,
                gender,
                birthdate,
                telephone,
                google_id,
                apple_id
            ) VALUES (?, ?, NULL, ?, ?, ?, ?, ?, ?, NULL)`, table)
		args = []interface{}{
			user.ID,
			user.EmailHash,
			user.LastName,
			user.FirstName,
			user.Gender,
			user.EncryptedBirthDate,
			user.Telephone,
			user.GoogleID,
		}
	case models.Apple:
		query = fmt.Sprintf(`
            INSERT INTO %s (
                id,
                email,
                password,
                lastname,
                firstname,
                gender,
                birthdate,
                telephone,
                google_id,
                apple_id
            ) VALUES (?, ?, NULL, ?, ?, ?, ?, ?, NULL, ?)`, table)
		args = []interface{}{
			user.ID,
			user.EmailHash,
			user.LastName,
			user.FirstName,
			user.Gender,
			user.EncryptedBirthDate,
			user.Telephone,
			user.AppleID,
		}
	case models.Mail:
		query = fmt.Sprintf(`
            INSERT INTO %s (
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
                address2,
                google_id,
                apple_id
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL)`, table)
		args = []interface{}{
			user.ID,
			user.EmailHash,
			user.Password,
			user.LastName,
			user.FirstName,
			user.Gender,
			user.EncryptedBirthDate,
			user.Telephone,
			user.PostalCode,
			user.City,
			user.Address1,
			user.Address2,
		}
	default:
		return rp.NewValidationErr(fmt.Errorf("unsupported provider type: %v", provider), "provider")
	}
	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, args...)
	} else {
		result, err = u.DB.ExecContext(ctx, query, args...)
	}
	if err != nil {
		return rp.NewNotCreatedErr(fmt.Errorf("failed to create user in %s table: %w", table, err), "pending user")
	}

	// Check if the insert was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(err, fmt.Sprintf("new user in %s table", table))
	}

	return nil
}
