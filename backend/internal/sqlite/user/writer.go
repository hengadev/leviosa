package user_repo

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
)

func (u *UserRepository) AddAccount(ctx context.Context, a *user.User) error {
	hashpassword, err := sqliteutil.HashPassword(a.Password)
	if err != nil {
		return err
	}
	_, err = u.DB.ExecContext(ctx, "INSERT INTO users (id, email, hashpassword, createdat, loggedinat, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", a.ID, a.Email, hashpassword, a.CreatedAt, a.LoggedInAt, a.Role, a.LastName, a.FirstName, a.Gender, a.BirthDate, a.Telephone, a.Address, a.City, a.PostalCard)
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}

func (u *UserRepository) ModifyAccount(ctx context.Context, user *user.User) error {
	query, fields := sqliteutil.WriteQueryUpdate(user)
	_, err := u.DB.ExecContext(ctx, query, fields...)
	if err != nil {
		return rp.NewRessourceUpdateErr(err)
	}
	return nil
}
