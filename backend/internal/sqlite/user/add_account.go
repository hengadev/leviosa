package userRepository

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
)

func (u *Repository) AddAccount(ctx context.Context, usr *userService.User) error {
	hashpassword, err := sqliteutil.HashPassword(usr.Password)
	if err != nil {
		return err
	}
	res, err := u.DB.ExecContext(ctx, "INSERT INTO users (email, password, createdat, loggedinat, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", usr.Email, hashpassword, usr.CreatedAt, usr.LoggedInAt, usr.Role, usr.LastName, usr.FirstName, usr.Gender, usr.BirthDate, usr.Telephone, usr.Address, usr.City, usr.PostalCard)
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	lastInsertID, err := res.LastInsertId()
	if lastInsertID == 0 {
		return fmt.Errorf("no user added")
	}
	return nil
}
