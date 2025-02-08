package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	testdb "github.com/GaryHY/leviosa/pkg/sqliteutil/testdatabase"
)

type sqliteRepository interface {
	GetDB() *sql.DB
}

type repoConstructor[T sqliteRepository] func(context.Context, *sql.DB) T

func recoverDB() {
	if r := recover(); r != nil {
		fmt.Println("In memory SQLite database setup failed.")
	}
}

type teardownFunc func()

func SetupRepository[T sqliteRepository](t testing.TB, ctx context.Context, version int64, constructor repoConstructor[T]) (T, teardownFunc) {
	t.Helper()
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
