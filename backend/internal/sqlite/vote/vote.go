package voteRepository

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type VoteRepository struct {
	DB *sql.DB
}

func (v *VoteRepository) GetDB() *sql.DB {
	return v.DB
}

func New(ctx context.Context, db *sql.DB) *VoteRepository {
	return &VoteRepository{db}
}
