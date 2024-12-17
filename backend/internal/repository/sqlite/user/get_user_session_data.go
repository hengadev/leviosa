package userRepository

import (
	"context"
	"database/sql"

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
	switch {
	case err == sql.ErrNoRows:
		return "", userService.UNKNOWN, rp.NewNotFoundError(err)
	case err != nil:
		return "", userService.UNKNOWN, rp.NewDatabaseErr(err)
	}
	return id, userService.ConvertToRole(role), nil
}
