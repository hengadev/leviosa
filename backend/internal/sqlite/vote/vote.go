package voteRepository

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type repository struct {
	DB *sql.DB
}

func (v *repository) GetDB() *sql.DB {
	return v.DB
}

func New(ctx context.Context, db *sql.DB) *repository {
	return &repository{db}
}
