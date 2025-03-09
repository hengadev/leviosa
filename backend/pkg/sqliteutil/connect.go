package sqliteutil

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func Connect(ctx context.Context, connStr string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping %q: %w", connStr, err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	return db, nil
}
