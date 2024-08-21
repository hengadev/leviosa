package userRepository

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
)

func (u *repository) AddAccount(ctx context.Context, usr *user.User) (int, error) {
	hashpassword, err := sqliteutil.HashPassword(usr.Password)
	if err != nil {
		return 0, err
	}
	res, err := u.DB.ExecContext(ctx, "INSERT INTO users (email, password, createdat, loggedinat, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", usr.Email, hashpassword, usr.CreatedAt, usr.LoggedInAt, usr.Role, usr.LastName, usr.FirstName, usr.Gender, usr.BirthDate, usr.Telephone, usr.Address, usr.City, usr.PostalCard)
	if err != nil {
		return 0, rp.NewRessourceCreationErr(err)
	}
	lastInsertID, err := res.LastInsertId()
	return int(lastInsertID), nil
}
