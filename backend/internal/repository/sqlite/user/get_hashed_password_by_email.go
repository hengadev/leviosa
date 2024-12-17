package userRepository

import (
	"context"
	"database/sql"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) GetHashedPasswordByEmail(ctx context.Context, usr *userService.Credentials) (string, error) {
	var hashedPassword string
	query := "SELECT password from users where email = ?;"
	err := u.DB.QueryRowContext(ctx, query, usr.Email).Scan(&hashedPassword)
	switch {
	case err == sql.ErrNoRows:
		return "", rp.NewNotFoundError(err, "user")
	case err != nil:
		return "", rp.NewDatabaseErr(err)
	}
	return hashedPassword, nil
}
