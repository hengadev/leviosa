package userRepository

import (
	"context"
	"errors"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// HasUser returns an error if no user is found in the users table with the specified hashed email.
func (u *Repository) HasUser(ctx context.Context, emailHash string) error {
	query := `
        SELECT EXISTS (
            SELECT 1 
            FROM users 
            WHERE email = ?
        );`
	var exists bool
	err := u.DB.QueryRowContext(ctx, query, emailHash).Scan(&exists)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}
	}
	if !exists {
		return rp.NewNotFoundErr(err, "user")
	}
	return nil
}
