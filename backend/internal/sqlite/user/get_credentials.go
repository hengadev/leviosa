package userRepository

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *Repository) GetCredentials(ctx context.Context, usr *userService.Credentials) (int, string, userService.Role, error) {
	var userRetrieved userService.User
	if err := u.DB.QueryRowContext(ctx, "SELECT id, password, role from users where email = ?;", usr.Email).Scan(
		&userRetrieved.ID,
		&userRetrieved.Password,
		&userRetrieved.Role,
	); err != nil {
		return 0, "", userService.ConvertToRole(""), rp.NewNotFoundError(err)
	}
	return userRetrieved.ID, userRetrieved.Password, userService.ConvertToRole(userRetrieved.Role), nil
}
