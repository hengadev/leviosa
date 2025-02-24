package sqliteutil

import (
	"context"
	"database/sql"
	"fmt"
)

func Connect(ctx context.Context, connStr string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping %q: %w", connStr, err)
	}
	return db, nil
}
