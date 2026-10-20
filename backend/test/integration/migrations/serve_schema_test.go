package migrations_test

import (
	"context"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Leviosa-care/leviosa/backend/internal/common/migrations"
)

// prefixFS returns an FS holding every embedded migration except the newest,
// standing in for the migration set an older release carried.
func prefixFS(t *testing.T) fs.FS {
	t.Helper()

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	if len(names) < 2 {
		t.Fatal("need at least two embedded migrations for the older-schema case")
	}
	sort.Strings(names)

	prefix := fstest.MapFS{}
	for _, name := range names[:len(names)-1] {
		data, err := migrations.FS.ReadFile(name)
		if err != nil {
			t.Fatalf("read embedded migration %s: %v", name, err)
		}
		prefix[name] = &fstest.MapFile{Data: data}
	}
	return prefix
}

func TestServeRefusesOlderSchema(t *testing.T) {
	dsn := createTestDB(t, "serve_older")
	ctx := context.Background()

	// An older release migrated this database: everything but the newest
	// migration is applied.
	db := openDB(t, dsn)
	if _, err := migrations.Apply(ctx, db, prefixFS(t)); err != nil {
		t.Fatalf("apply the older migration set: %v", err)
	}
	older, err := migrations.CurrentVersion(ctx, db)
	if err != nil {
		t.Fatalf("read database version: %v", err)
	}

	run := runApp(t, "serve", serveEnv(dsn, freePort(t)))
	if run.exitCode == 0 {
		t.Fatalf("expected serve to exit non-zero against an older schema; output:\n%s", run.output)
	}

	// The refusal must name both versions so the skipped step is obvious.
	for _, want := range []string{
		strconv.FormatInt(older, 10),
		strconv.FormatInt(migrations.NewestVersion(), 10),
		"out of date",
	} {
		if !strings.Contains(run.output, want) {
			t.Fatalf("expected the refusal to mention %q; output:\n%s", want, run.output)
		}
	}
}

func TestServeStartsAgainstCurrentSchema(t *testing.T) {
	dsn := createTestDB(t, "serve_current")
	ctx := context.Background()

	db := openDB(t, dsn)
	if _, err := migrations.ApplyAll(ctx, db); err != nil {
		t.Fatalf("apply embedded migrations: %v", err)
	}

	proc := startServe(t, dsn, freePort(t))
	proc.stop(t)

	if code := proc.cmd.ProcessState.ExitCode(); code != 0 {
		t.Fatalf("expected a clean exit, got %d; output:\n%s", code, proc.output.String())
	}
	if !strings.Contains(proc.output.String(), "Server stopped gracefully") {
		t.Fatalf("expected a graceful shutdown log; output:\n%s", proc.output.String())
	}
}

func TestServeStartsAgainstNewerSchema(t *testing.T) {
	dsn := createTestDB(t, "serve_newer")
	ctx := context.Background()

	// A rollback: this binary is older than the release that migrated the
	// database. serve must accept the newer schema and keep serving.
	db := openDB(t, dsn)
	if _, err := migrations.ApplyAll(ctx, db); err != nil {
		t.Fatalf("apply embedded migrations: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, true)",
		migrations.NewestVersion()+1,
	); err != nil {
		t.Fatalf("simulate a newer schema version: %v", err)
	}

	proc := startServe(t, dsn, freePort(t))
	proc.stop(t)

	if code := proc.cmd.ProcessState.ExitCode(); code != 0 {
		t.Fatalf("expected a clean exit, got %d; output:\n%s", code, proc.output.String())
	}
	if !strings.Contains(proc.output.String(), "Server stopped gracefully") {
		t.Fatalf("expected a graceful shutdown log; output:\n%s", proc.output.String())
	}
}
