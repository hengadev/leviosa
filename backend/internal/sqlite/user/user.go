package user_repo

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

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(ctx context.Context) (*UserRepository, error) {
	connStr := os.Getenv("userdb")
	db, err := sqliteutil.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	store := &UserRepository{db}
	adminpassword, err := sqliteutil.HashPassword("1234")
	if err != nil {
		return nil, err
	}
	queries := []string{
		"CREATE TABLE IF NOT EXISTS users (id UUID NOT NULL, email TEXT NOT NULL PRIMARY KEY, hashpassword TEXT NOT NULL, createdat TEXT NOT NULL, loggedinat TEXT NOT NULL, role TEXT NOT NULL, lastname TEXT NOT NULL, firstname TEXT NOT NULL, gender TEXT NOT NULL, birthdate TEXT NOT NULL, telephone TEXT NOT NULL, address TEXT NOTN NULL, city TEXT NOT NULL, postalcard INTEGER NOT NULL);",
		// The admin user for testing purposes.
		fmt.Sprintf("INSERT INTO users VALUES (3439434532245, 'test@example.fr', '%s', 'admin', 'HENRY', 'Livio', 'male', '20/08/1999', '0000 00 00 00', 'admin address', 'admin city', 'admin postalcard');", adminpassword),
	}
	sqliteutil.Init(store.DB, queries...)
	return store, nil
}

// TODO: use transaction if need be.
