package userRepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) HasOAuthUser(ctx context.Context, emailHash string, p models.ProviderType) error {
	tx, err := u.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return rp.NewDatabaseErr(fmt.Errorf("failed to start transaction: %w", err))
	}
	defer tx.Rollback()
	// this is the string provider here
	provider := string(p)
	tables := []string{"users", "pending_users"}
	for _, table := range tables {
		query := fmt.Sprintf(`
        SELECT EXISTS (
            SELECT 1 
            FROM %s 
            WHERE email = ? 
            AND %s_id IS NOT NULL
            AND %s_id != ''
        );`, table, provider, provider)
		var exists bool
		err = tx.QueryRowContext(ctx, query, emailHash).Scan(&exists)
		if err != nil {
			switch {
			case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
				return rp.NewContextErr(err)
			default:
				return rp.NewDatabaseErr(err)
			}
		}
		if !exists {
			return rp.NewNotFoundErr(err, "google account with specified email hash")
		}
	}

	if err := tx.Commit(); err != nil {
		return rp.NewDatabaseErr(fmt.Errorf("commit transaction: %w", err))
	}
	return nil
}
