package userRepository

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type UserRepository struct {
	DB *sql.DB
}

func (u *UserRepository) GetDB() *sql.DB {
	return u.DB
}

func New(ctx context.Context, db *sql.DB) *UserRepository {
	return &UserRepository{db}
}
