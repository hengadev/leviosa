package carePlanRepository

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) GetDB() *sql.DB {
	return r.DB
}

func New(ctx context.Context, db *sql.DB) *Repository {
	return &Repository{db}
}
