package testdatabase

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/sqlite"
	util "github.com/GaryHY/event-reservation-app/pkg/sqliteutil"
	"github.com/pressly/goose/v3"
)

func NewDatabase(ctx context.Context) (*sql.DB, error) {
	db, err := util.Connect(ctx, ":memory:")
	if err != nil {
		return nil, fmt.Errorf("create new in memory database: %w", err)
	}
	return db, nil
}

func Teardown(db *sql.DB) error {
	return db.Close()
}

func Setup(ctx context.Context, db *sql.DB, version int64) error {
	// setup goose
	goose.SetBaseFS(nil)
	// Set the dialect to SQLite3
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("Failed to set dialect: %w", err)
	}
	if err := goose.UpToContext(ctx, db, os.Getenv("TEST_MIGRATION_PATH"), version); err != nil {
		return fmt.Errorf("test migration up failed with version %d: %w", version, err)
	}
	return nil
}

type RepoConstructor[T sqlite.Repository] func(context.Context, *sql.DB) T

func SetupRepo[T sqlite.Repository](ctx context.Context, version int64, constructor RepoConstructor[T]) (T, error) {
	var repo T
	db, err := NewDatabase(ctx)
	if err != nil {
		return repo, fmt.Errorf("database connection: %s", err)
	}
	repo = constructor(ctx, db)
	if err := Setup(ctx, repo.GetDB(), version); err != nil {
		return repo, fmt.Errorf("migration to the database: %s", err)
	}
	return repo, nil
}
