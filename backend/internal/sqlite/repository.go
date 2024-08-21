package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	testdb "github.com/GaryHY/event-reservation-app/pkg/sqliteutil/testdatabase"
)

type repository interface {
	GetDB() *sql.DB
}

type repoConstructor[T repository] func(context.Context, *sql.DB) T

func recoverDB() {
	if r := recover(); r != nil {
		fmt.Println("In memory SQLite database setup failed.")
	}
}

type teardownFunc func()

func SetupRepository[T repository](t testing.TB, ctx context.Context, version int64, constructor repoConstructor[T]) (T, teardownFunc) {
	var repo T
	db, err := testdb.NewDatabase(ctx)
	if err != nil {
		t.Errorf("database connection: %s", err)
	}
	repo = constructor(ctx, db)
	if err := testdb.Setup(ctx, repo.GetDB(), version); err != nil {
		t.Errorf("database setup: %s", err)
	}
	teardown := func() {
		defer recoverDB()
		err := db.Close()
		if err != nil {
			t.Errorf("database teardown: %s", err)
		}
	}
	return repo, teardown
}
