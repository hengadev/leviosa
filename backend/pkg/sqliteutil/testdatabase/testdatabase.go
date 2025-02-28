package testdatabase

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	util "github.com/hengadev/leviosa/pkg/sqliteutil"
	"github.com/pressly/goose/v3"
)

func NewDatabase(ctx context.Context) (*sql.DB, error) {
	db, err := util.Connect(ctx, ":memory:")
	if err != nil {
		return nil, fmt.Errorf("create new in memory database: %w", err)
	}
	return db, nil
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
