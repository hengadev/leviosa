package sqliteutil

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
)

func MigrateFS(ctx context.Context, db *sql.DB, fsys fs.FS) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	migrationsDir, err := fs.Sub(fsys, "migrations")
	if err != nil {
		return fmt.Errorf("sub filesystem: %w", err)
	}
	migrationFiles, err := fs.ReadDir(migrationsDir, ".")
	fmt.Printf("the sql directory : %q\n", migrationsDir)
	queries := make([]string, len(migrationFiles))
	for i, file := range migrationFiles {
		filename := file.Name()
		fmt.Printf("In file number %d, doing the migration for the file : %s!\n", i+1, filename)
		if file.IsDir() {
			return fmt.Errorf("want file; got directory: %q", err)
		}
		if filename[:4] != fmt.Sprintf("%04d", i+1) {
			return fmt.Errorf("file should start with %04d, got %q", i+1, filename[:4])
		}
		b, err := fs.ReadFile(migrationsDir, filename)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		queries[i] = string(b)
	}
	if err := migrate(ctx, tx, queries); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func migrate(ctx context.Context, tx *sql.Tx, migrations []string) error {
	defer tx.Rollback()
	fail := func(err error) error {
		return fmt.Errorf("migration: %w", err)
	}
	if len(migrations) == 0 {
		return nil
	}
	for _, stmt := range migrations {
		if stmt = strings.TrimSpace(stmt); stmt == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fail(err)
	}
	return nil
}
