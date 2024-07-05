package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/sqliteutil"

	_ "github.com/mattn/go-sqlite3"
)

// general
type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(ctx context.Context) (*UserRepository, error) {
	connStr := os.Getenv("userdb")
	db, err := sqliteutil.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	// TODO: initialise the admin if the env variable is set to dev.
	// or maybe us flags for this ?
	if os.Getenv("env") == "dev" {
		ProdInit(db)
	}
	return &UserRepository{db}, nil
}

// Cette fonction je ne l'utilise qu'en environment de production dev, comment le retranscrire dans mon code ?
func ProdInit(db *sql.DB) error {
	// init some administrator
	adminpassword, err := sqliteutil.HashPassword("1234")
	if err != nil {
		return err
	}
	queries := []string{
		"CREATE TABLE IF NOT EXISTS users (id UUID NOT NULL, email TEXT NOT NULL PRIMARY KEY, hashpassword TEXT NOT NULL, createdat TEXT NOT NULL, loggedinat TEXT NOT NULL, role TEXT NOT NULL, lastname TEXT NOT NULL, firstname TEXT NOT NULL, gender TEXT NOT NULL, birthdate TEXT NOT NULL, telephone TEXT NOT NULL, address TEXT NOTN NULL, city TEXT NOT NULL, postalcard INTEGER NOT NULL);",
		// The admin user for testing purposes.
		fmt.Sprintf("INSERT INTO users VALUES (3439434532245, 'test@example.fr', '%s', 'admin', 'HENRY', 'Livio', 'male', '20/08/1999', '0000 00 00 00', 'admin address', 'admin city', 'admin postalcard');", adminpassword),
	}
	sqliteutil.Init(db, queries...)
	return nil
}

// reader
// here put the function that you need to put brother
func (u *UserRepository) FindAccountByID(ctx context.Context, id string) (*user.User, error) {
	var user *user.User
	if err := u.DB.QueryRowContext(ctx, "select (id, lastname, firstname, gender, birthdate, telephone, address, city, postalcard) from users where id = ?;", id).Scan(user); err != nil {
		return user, rp.NewNotFoundError(err)
	}
	return user, nil
}

func (u *UserRepository) GetUserIDBySessionID(ctx context.Context, sessionID string) (id string) {
	u.DB.QueryRowContext(ctx, "SELECT userid from sessions where id = ?;", sessionID).Scan(&id)
	return
}

func (u *UserRepository) GetAllUsers(ctx context.Context) ([]*user.User, error) {
	users := make([]*user.User, 0)
	rows, err := u.DB.QueryContext(ctx, "SELECT * FROM users;")
	if err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	defer rows.Close()

	for rows.Next() {
		user := &user.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
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
func (u *UserRepository) AddAccount(ctx context.Context, usr *user.User) (string, error) {
	hashpassword, err := sqliteutil.HashPassword(usr.Password)
	if err != nil {
		return "", err
	}
	_, err = u.DB.ExecContext(ctx, "INSERT INTO users (id, email, hashpassword, createdat, loggedinat, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", usr.ID, usr.Email, hashpassword, usr.CreatedAt, usr.LoggedInAt, usr.Role, usr.LastName, usr.FirstName, usr.Gender, usr.BirthDate, usr.Telephone, usr.Address, usr.City, usr.PostalCard)
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	return usr.ID, nil
}

func (u *UserRepository) ModifyAccount(ctx context.Context, user *user.User) error {
	query, fields := sqliteutil.WriteUpdateQuery(user)
	_, err := u.DB.ExecContext(ctx, query, fields...)
	if err != nil {
		return rp.NewRessourceUpdateErr(err)
	}
	return nil
}

func (u *UserRepository) DeleteUser(userID string) (string, error) {
	_, err := u.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return "", rp.NewRessourceDeleteErr(err)
	}
	return userID, nil
}

func (u *UserRepository) UpdateUser(ctx context.Context, user *user.User) (string, error) {
	query, fields := sqliteutil.WriteUpdateQuery(user)
	_, err := u.DB.ExecContext(
		ctx,
		query,
		fields...,
	)

	if err != nil {
		return "", rp.NewRessourceUpdateErr(err)
	}
	return user.ID, nil
}
