package userRepository

import (
	"context"
	"errors"

	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (u *Repository) DeleteUser(ctx context.Context, userID string) error {
	result, err := u.DB.ExecContext(ctx, "DELETE FROM users WHERE id = ?;", userID)
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
		return rp.NewNotDeletedErr(err, "user")
	}
	return nil
}
