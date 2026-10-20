package migrations_test

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Leviosa-care/leviosa/backend/internal/common/migrations"
)

func TestMigrateModeAppliesAllMigrationsToEmptyDatabase(t *testing.T) {
	dsn := createTestDB(t, "migrate_empty")
	ctx := context.Background()

	run := runApp(t, "migrate", baseEnv(dsn, nil))
	if run.exitCode != 0 {
		t.Fatalf("expected exit code 0 on an empty database, got %d; output:\n%s", run.exitCode, run.output)
	}

	db := openDB(t, dsn)
	version, err := migrations.CurrentVersion(ctx, db)
	if err != nil {
		t.Fatalf("read database version: %v", err)
	}
	if version != migrations.NewestVersion() {
		t.Fatalf("expected database at the newest embedded version %d, got %d", migrations.NewestVersion(), version)
	}

	// Running it again must be a successful no-op, not an error.
	run = runApp(t, "migrate", baseEnv(dsn, nil))
	if run.exitCode != 0 {
		t.Fatalf("expected exit code 0 on an already-migrated database, got %d; output:\n%s", run.exitCode, run.output)
	}
}

// firstMigration is the oldest embedded migration. It creates catalog.status
// without IF NOT EXISTS, which the broken-migration case below relies on.
const firstMigration = "20250602201515_catalog_init_schema.sql"

func TestMigrateModeExitsNonZeroOnBrokenMigration(t *testing.T) {
	dsn := createTestDB(t, "migrate_broken")
	ctx := context.Background()

	// Make a real embedded migration fail, the way a bad release fails
	// mid-deploy: a pre-existing catalog.status type breaks the first one.
	db := openDB(t, dsn)
	for _, stmt := range []string{
		"CREATE SCHEMA catalog",
		"CREATE TYPE catalog.status AS ENUM ('draft')",
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("prepare a conflicting schema (%s): %v", stmt, err)
		}
	}

	run := runApp(t, "migrate", baseEnv(dsn, nil))
	if run.exitCode == 0 {
		t.Fatalf("expected a non-zero exit code on a failing migration; output:\n%s", run.output)
	}
	if !strings.Contains(run.output, firstMigration) {
		t.Fatalf("expected the output to name the failing migration %s; output:\n%s", firstMigration, run.output)
	}

	// The failed migration must not be recorded as applied.
	version, err := migrations.CurrentVersion(ctx, db)
	if err != nil {
		t.Fatalf("read database version: %v", err)
	}
	if version != 0 {
		t.Fatalf("a failed migration must not advance the version: expected 0, got %d", version)
	}
}

func TestMigrateModeExitsNonZeroWhenDatabaseUnreachable(t *testing.T) {
	run := runApp(t, "migrate", baseEnv("postgresql://nope@127.0.0.1:1/none", nil))
	if run.exitCode == 0 {
		t.Fatalf("expected a non-zero exit code, got 0; output:\n%s", run.output)
	}
}

func TestMigrateModeAcceptsNewerSchema(t *testing.T) {
	dsn := createTestDB(t, "migrate_newer")
	ctx := context.Background()

	// A rollback deploy runs the older release's migrate before its serve,
	// against a schema the newer release already migrated: that must be a
	// successful no-op, leaving the newer version in place.
	db := openDB(t, dsn)
	if _, err := migrations.ApplyAll(ctx, db); err != nil {
		t.Fatalf("apply embedded migrations: %v", err)
	}
	newer := simulateNewerSchema(t, db)

	run := runApp(t, "migrate", baseEnv(dsn, nil))
	if run.exitCode != 0 {
		t.Fatalf("expected exit code 0 against a newer schema, got %d; output:\n%s", run.exitCode, run.output)
	}

	version, err := migrations.CurrentVersion(ctx, db)
	if err != nil {
		t.Fatalf("read database version: %v", err)
	}
	if version != newer {
		t.Fatalf("expected the newer version %d to stay in place, got %d", newer, version)
	}
}

func TestApplyBrokenMigrationFailsAndNamesIt(t *testing.T) {
	dsn := createTestDB(t, "broken_migration")
	ctx := context.Background()

	db := openDB(t, dsn)
	if _, err := migrations.ApplyAll(ctx, db); err != nil {
		t.Fatalf("apply embedded migrations: %v", err)
	}

	// A migration set whose newest entry is broken, standing in for a bad
	// release: version far above the embedded set, SQL that cannot run.
	broken := fstest.MapFS{
		"20990101000000_broken_migration.sql": &fstest.MapFile{Data: []byte(`-- +goose Up
-- +goose StatementBegin
THIS IS NOT VALID SQL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
`)},
	}

	_, err := migrations.Apply(ctx, db, broken)
	if err == nil {
		t.Fatal("expected the broken migration to fail")
	}
	if !strings.Contains(err.Error(), "20990101000000_broken_migration.sql") {
		t.Fatalf("expected the error to name the broken migration file, got: %v", err)
	}

	// The failed migration must not be recorded as applied.
	version, err := migrations.CurrentVersion(ctx, db)
	if err != nil {
		t.Fatalf("read database version: %v", err)
	}
	if version != migrations.NewestVersion() {
		t.Fatalf("a failed migration must not advance the version: expected %d, got %d", migrations.NewestVersion(), version)
	}
}
