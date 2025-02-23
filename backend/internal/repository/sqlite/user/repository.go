package userRepository

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	DB *sql.DB
}

func (u *Repository) GetDB() *sql.DB {
	return u.DB
}

func New(ctx context.Context, db *sql.DB) *Repository {
	return &Repository{db}
}
