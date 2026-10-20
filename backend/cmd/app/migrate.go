package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/app"
	"github.com/Leviosa-care/leviosa/backend/internal/common/migrations"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for the migration connection
)

// migrate applies pending embedded migrations and exits: `leviosa migrate`.
// Deploys run it as a one-off before the new backend container starts, so a
// failed migration stops the deploy instead of crash-looping the app. It
// exits non-zero and names the migration that failed. See the migration rule
// in backend/CLAUDE.md.
func migrate(ctx context.Context, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("migrate takes no arguments, got %q", args)
	}

	logger := setupLogger()

	cfg, err := app.LoadConfig(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := sql.Open("pgx", cfg.PostgresURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	before, err := migrations.CurrentVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("read database migration version: %w", err)
	}

	after, err := migrations.ApplyAll(ctx, db)
	if err != nil {
		return err // already names the failing migration
	}

	logger.InfoContext(ctx, "database migrated",
		"from_version", before,
		"to_version", after,
	)
	return nil
}
