package userRepository

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (u *UserRepository) GetCredentials(ctx context.Context, usr *user.Credentials) (int, string, user.Role, error) {
	var userRetrieved user.User
	if err := u.DB.QueryRowContext(ctx, "SELECT id, password, role from users where email = ?;", usr.Email).Scan(
		&userRetrieved.ID,
		&userRetrieved.Password,
		&userRetrieved.Role,
	); err != nil {
		return 0, "", user.ConvertToRole(""), rp.NewNotFoundError(err)
	}
	return userRetrieved.ID, userRetrieved.Password, user.ConvertToRole(userRetrieved.Role), nil
}
