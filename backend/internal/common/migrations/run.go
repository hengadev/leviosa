package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/pressly/goose/v3"
)

// gooseTableName is goose's default version table. The raw queries in
// CurrentVersion must stay in sync with the store the provider below uses.
const gooseTableName = "goose_db_version"

// newProvider builds a goose provider for fsys. The provider API is used
// instead of the package-level goose.SetBaseFS/SetDialect globals so several
// migration sets (the embedded one and test-only ones) can coexist in one
// process without fighting over global state.
func newProvider(db *sql.DB, fsys fs.FS) (*goose.Provider, error) {
	p, err := goose.NewProvider(goose.DialectPostgres, db, fsys)
	if err != nil {
		return nil, fmt.Errorf("create goose provider: %w", err)
	}
	return p, nil
}

// Apply runs every pending migration found in fsys and returns the database's
// migration version afterwards. A failing migration is reported by filename.
func Apply(ctx context.Context, db *sql.DB, fsys fs.FS) (int64, error) {
	p, err := newProvider(db, fsys)
	if err != nil {
		return 0, err
	}

	if _, err := p.Up(ctx); err != nil {
		var partial *goose.PartialError
		if errors.As(err, &partial) && partial.Failed != nil && partial.Failed.Source != nil {
			return 0, fmt.Errorf("migration %s failed: %w", partial.Failed.Source.Path, partial.Err)
		}
		return 0, fmt.Errorf("apply migrations: %w", err)
	}

	return CurrentVersion(ctx, db)
}

// ApplyAll runs every pending migration embedded in the binary. It is what
// the migrate mode runs; serve never calls it (it only checks the version).
func ApplyAll(ctx context.Context, db *sql.DB) (int64, error) {
	return Apply(ctx, db, FS)
}

// CurrentVersion returns the database's migration version: the highest
// version goose has recorded, or 0 for a database that was never migrated.
func CurrentVersion(ctx context.Context, db *sql.DB) (int64, error) {
	// goose's own getters error on a never-migrated database ("missing zero
	// version migration"); callers here need a plain 0 instead, so the
	// version table is read directly, mirroring goose's max(version_id).
	var exists bool
	if err := db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT to_regclass('%s') IS NOT NULL", gooseTableName),
	).Scan(&exists); err != nil {
		return 0, fmt.Errorf("check %s table: %w", gooseTableName, err)
	}
	if !exists {
		return 0, nil
	}

	var version sql.NullInt64
	if err := db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT max(version_id) FROM %s", gooseTableName),
	).Scan(&version); err != nil {
		return 0, fmt.Errorf("read %s table: %w", gooseTableName, err)
	}
	if !version.Valid {
		return 0, nil
	}
	return version.Int64, nil
}

// NewestVersion returns the highest migration version embedded in the binary.
func NewestVersion() int64 {
	v, err := newestVersion(FS)
	if err != nil {
		// The embedded set is compiled in; an empty or unreadable one is a
		// build bug, not a runtime condition.
		panic(fmt.Sprintf("migrations: embedded set invalid: %v", err))
	}
	return v
}

// newestVersion returns the highest {version}_*.sql version in fsys. Files
// without a numeric version prefix are skipped, matching what goose collects.
func newestVersion(fsys fs.FS) (int64, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return 0, fmt.Errorf("read migrations: %w", err)
	}

	var newest int64
	found := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		version, err := parseVersion(e.Name())
		if err != nil {
			continue
		}
		if !found || version > newest {
			newest, found = version, true
		}
	}
	if !found {
		return 0, errors.New("no .sql migrations found")
	}
	return newest, nil
}

func parseVersion(name string) (int64, error) {
	prefix, _, _ := strings.Cut(name, "_")
	version, err := strconv.ParseInt(prefix, 10, 64)
	if err != nil || version <= 0 {
		return 0, fmt.Errorf("missing or invalid version prefix")
	}
	return version, nil
}

// EnsureCurrent verifies the database's migration version is not older than
// the newest embedded migration, so serve refuses to start when the migrate
// step was skipped. A higher version — rolling back to an older release — is
// accepted, because migrations are never reverted (see the migration rule in
// backend/CLAUDE.md).
func EnsureCurrent(ctx context.Context, db *sql.DB) error {
	current, err := CurrentVersion(ctx, db)
	if err != nil {
		return err
	}

	newest := NewestVersion()
	if current < newest {
		return fmt.Errorf(
			"database schema is out of date: version %d is older than the newest embedded migration %d — run the migrate mode before serving",
			current, newest,
		)
	}
	return nil
}
