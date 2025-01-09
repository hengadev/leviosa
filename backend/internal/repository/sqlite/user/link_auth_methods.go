package userRepository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (u *Repository) linkAuthMethod(ctx context.Context, tx *sql.Tx, userID string, user *models.User, provider models.ProviderType, table string) error {
	switch provider {
	case models.Google:
		// Add Apple authentication
		_, err := tx.ExecContext(ctx,
			fmt.Sprintf(`UPDATE %s SET 
                google_id = ?,
            WHERE id = ?`, table),
			user.GoogleID,
			userID,
		)
		return err
	case models.Apple:
		// Add Apple authentication
		_, err := tx.ExecContext(ctx,
			fmt.Sprintf(`UPDATE %s SET 
                apple_id = ?,
            WHERE id = ?`, table),
			user.AppleID,
			userID,
		)
		return err
	case models.Mail:
		// Add email authentication
		_, err := tx.ExecContext(ctx,
			fmt.Sprintf(`UPDATE %s SET 
                password_hash = ?,
            WHERE id = ?`, table),
			user.PasswordHash,
			userID,
		)
		return err
	default:
		return rp.NewValidationErr(fmt.Errorf("unsupported provider type: %v", provider), "provider")
	}
}
