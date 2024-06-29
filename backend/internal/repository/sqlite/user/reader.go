package user_repo

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

// here put the functiono that you need to put brother
func (u *UserRepository) FindAccountByID(ctx context.Context, id string) (*user.User, error) {
	var user *user.User
	if err := u.DB.QueryRowContext(ctx, "select (id, email, hashpassword, createdat, loggedinat, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard) from users where id = ?;", id).Scan(user); err != nil {
		return user, rp.NewNotFoundError(err)
	}
	return user, nil
}
