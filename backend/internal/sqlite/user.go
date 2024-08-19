package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"

	_ "github.com/mattn/go-sqlite3"
)

// general
type UserRepository struct {
	DB *sql.DB
}

func (u *UserRepository) GetDB() *sql.DB {
	return u.DB
}

func NewUserRepository(ctx context.Context, db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

// reader
func (u *UserRepository) FindAccountByID(ctx context.Context, id int) (*user.User, error) {
	var user user.User
	if err := u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?;", id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.LoggedInAt,
		&user.Role,
		&user.BirthDate,
		&user.LastName,
		&user.FirstName,
		&user.Gender,
		&user.Telephone,
		&user.Address,
		&user.City,
		&user.PostalCard,
	); err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	return &user, nil
}

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

func (u *UserRepository) GetAllUsers(ctx context.Context) ([]*user.User, error) {
	var users []*user.User
	query := "SELECT email, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard FROM users;"
	rows, err := u.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	defer rows.Close()

	for rows.Next() {
		user := &user.User{}
		err := rows.Scan(
			&user.Email,
			&user.Role,
			&user.LastName,
			&user.FirstName,
			&user.Gender,
			&user.BirthDate,
			&user.Telephone,
			&user.Address,
			&user.City,
			&user.PostalCard,
		)
		if err != nil {
			return nil, rp.NewNotFoundError(err)
		}
		users = append(users, user)
	}
	return users, nil
}

// writer
func (u *UserRepository) AddAccount(ctx context.Context, usr *user.User) (int, error) {
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

func (u *UserRepository) ModifyAccount(
	ctx context.Context,
	user *user.User,
	whereMap map[string]any,
	prohibitedFields ...string,
) (int, error) {
	fail := func(err error) (int, error) {
		return 0, rp.NewRessourceUpdateErr(err)
	}
	if user == nil {
		return fail(fmt.Errorf("nil user"))
	}
	query, values, err := sqliteutil.WriteUpdateQuery(*user, whereMap, prohibitedFields...)
	if err != nil {
		return fail(err)
	}
	res, err := u.DB.ExecContext(ctx, query, values...)
	if err != nil {
		return fail(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fail(err)
	}
	return int(rowsAffected), nil
}

func (u *UserRepository) DeleteUser(ctx context.Context, userID int) (int, error) {
	res, err := u.DB.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return 0, rp.NewRessourceDeleteErr(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, rp.NewRessourceDeleteErr(err)
	}
	return int(rowsAffected), nil
}
