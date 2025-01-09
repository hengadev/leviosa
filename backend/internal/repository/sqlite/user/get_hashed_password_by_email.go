package userRepository

import (
	"context"
	"database/sql"
	"errors"

	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (u *Repository) GetHashedPasswordByEmail(ctx context.Context, email string) (string, error) {
	var hashedPassword string
	query := "SELECT password from users where email = ?;"
	err := u.DB.QueryRowContext(ctx, query, email).Scan(&hashedPassword)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", rp.NewNotFoundErr(err, "user")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return "", rp.NewContextErr(err)
		default:
			return "", rp.NewDatabaseErr(err)
		}
	}
	return hashedPassword, nil
}
