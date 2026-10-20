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

func TestMigrateModeExitsNonZeroOnFailure(t *testing.T) {
	// An unreachable database is the simplest failing migration: the mode
	// must exit non-zero rather than pretend it worked.
	run := runApp(t, "migrate", baseEnv("postgresql://nope@127.0.0.1:1/none", nil))
	if run.exitCode == 0 {
		t.Fatalf("expected a non-zero exit code, got 0; output:\n%s", run.output)
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
