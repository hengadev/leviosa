package sqlite_test

// what I need to test in the sqlite_test package

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
)

var johndoe = &user.User{
	ID:         1,
	Email:      "john.doe@gmail.com",
	Password:   "$a9rfNhA$N$A78#m",
	CreatedAt:  time.Now().Add(-time.Hour * 4),
	LoggedInAt: time.Now().Add(-time.Hour * 4),
	Role:       user.BASIC.String(),
	BirthDate:  "1998-07-12",
	LastName:   "DOE",
	FirstName:  "John",
	Gender:     "M",
	Telephone:  "0123456789",
	Address:    "Impasse Inconnue",
	City:       "Paris",
	PostalCard: 12345,
}

type RepoConstructor[T sqlite.Repository] func(context.Context, *sql.DB) T

func setupRepo[T sqlite.Repository](ctx context.Context, version int64, constructor RepoConstructor[T]) (T, error) {
	var repo T
	db, err := testdb.NewDatabase(ctx)
	if err != nil {
		return repo, fmt.Errorf("database connection: %s", err)
	}
	repo = constructor(ctx, db)
	if err := testdb.Setup(ctx, repo.GetDB(), version); err != nil {
		return repo, fmt.Errorf("migration to the database: %s", err)
	}
	return repo, nil
}

func getOnlyUser(ctx context.Context, db *sql.DB) (*user.User, error) {
	var foundUser user.User
	if err := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = 1").Scan(
		&foundUser.ID,
		&foundUser.Email,
		&foundUser.Password,
		&foundUser.CreatedAt,
		&foundUser.LoggedInAt,
		&foundUser.Role,
		&foundUser.BirthDate,
		&foundUser.LastName,
		&foundUser.FirstName,
		&foundUser.Gender,
		&foundUser.Telephone,
		&foundUser.Address,
		&foundUser.City,
		&foundUser.PostalCard,
	); err != nil {
		return nil, fmt.Errorf("user not found after addition to database: %s", err)
	}
	return &foundUser, nil
}
