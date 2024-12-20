package userRepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) GetUserSessionData(ctx context.Context, email string) (string, userService.Role, error) {
	var id, role string
	query := "SELECT id, role from users where email = ?;"
	err := u.DB.QueryRowContext(ctx, query, email).Scan(
		&id,
		&role,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", userService.UNKNOWN, rp.NewNotFoundError(err, "user session data")
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return "", userService.UNKNOWN, rp.NewContextError(err)
		default:
			return "", userService.UNKNOWN, rp.NewDatabaseErr(err)
		}
	}
	return id, userService.ConvertToRole(role), nil
}
