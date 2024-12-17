package userRepository

import (
	"context"
	"database/sql"

	// "github.com/GaryHY/event-reservation-app/internal/domain/user"

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

// type readerRepository struct {
// 	DB *sql.DB
// }
//
// func NewReaderRepository(ctx context.Context, db *sql.DB) *readerRepository {
// 	return &readerRepository{db}
// }
//
// type repository struct {
// 	*readerRepository
// }
//
// func New(ctx context.Context, rp *readerRepository) *repository {
// 	return &repository{rp}
// }
