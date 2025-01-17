package sqliteutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/GaryHY/leviosa/pkg/flags"

	"github.com/pressly/goose/v3"
)

// MigrationConfig holds the configuration for database migrations
type MigrationConfig struct {
	DB           *sql.DB
	Env          mode.EnvMode
	MigrationDir string
}

// targetVersions maps environment modes to their target migration versions
var targetVersions = map[mode.EnvMode]int64{
	mode.ModeDev:     0,
	mode.ModeStaging: 20241228203428,
	mode.ModeProd:    20241228203428,
}

// NewMigrationConfig creates a new migration configuration with the specified settings.
//
// Parameters:
//   - db: A sql.DB pointer representing the database connection.
//   - env: An EnvMode value specifying the environment (dev, staging, prod).
//
// Returns:
//   - *MigrationConfig: A pointer to the created configuration if successful.
//   - error: An error if MIGRATION_PATH environment variable is not set.
func NewMigrationConfig(db *sql.DB, env mode.EnvMode) (*MigrationConfig, error) {
	migrationDir := os.Getenv("MIGRATION_PATH")
	if migrationDir == "" {
		return nil, fmt.Errorf("MIGRATION_PATH environment variable is required")
	}

	return &MigrationConfig{
		DB:           db,
		Env:          env,
		MigrationDir: migrationDir,
	}, nil
}

// SetMigrations configures and executes database migrations for the specified environment.
//
// Parameters:
//   - ctx: A context.Context instance to manage migration lifecycle and cancellation.
//   - cfg: A MigrationConfig pointer containing database, environment and directory settings.
//
// Returns:
//   - error: An error if migration setup fails, migrations cannot be executed, or environment
//     is not supported. Returns nil on successful migration.
func SetMigrations(ctx context.Context, cfg *MigrationConfig) error {
	if cfg == nil {
		return fmt.Errorf("migration config cannot be nil")
	}

	if err := initializeGoose(); err != nil {
		return fmt.Errorf("initializing goose: %w", err)
	}

	targetVersion, ok := targetVersions[cfg.Env]
	if !ok {
		return fmt.Errorf("unsupported environment for migration: %s", cfg.Env.String())
	}

	if err := runMigrations(ctx, cfg, targetVersion); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

// initializeGoose sets up the basic Goose migration configuration.
//
// Returns:
//   - error: An error if setting the SQLite dialect fails. Returns nil on successful setup.
func initializeGoose() error {
	goose.SetBaseFS(nil)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("setting dialect: %w", err)
	}

	return nil
}

// runMigrations executes database migrations up to the specified version.
//
// Parameters:
//   - ctx: A context.Context instance to manage migration lifecycle and cancellation.
//   - cfg: A MigrationConfig pointer containing database and directory settings.
//   - targetVersion: An int64 representing the desired migration version. Use 0 for latest.
//
// Returns:
//   - error: An error if migrations fail to execute. Returns nil on successful migration.
func runMigrations(ctx context.Context, cfg *MigrationConfig, targetVersion int64) error {
	if targetVersion == 0 {
		if err := goose.UpContext(ctx, cfg.DB, cfg.MigrationDir); err != nil {
			return fmt.Errorf("running all migrations: %w", err)
		}
		return nil
	}

	if err := goose.UpToContext(ctx, cfg.DB, cfg.MigrationDir, targetVersion); err != nil {
		return fmt.Errorf("running migrations up to version %d: %w", targetVersion, err)
	}

	return nil
}
