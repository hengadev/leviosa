package userRepository

import (
	"context"
	"database/sql"

	// "github.com/GaryHY/event-reservation-app/internal/domain/user"

	_ "github.com/mattn/go-sqlite3"
)

type repository struct {
	DB *sql.DB
}

func (u *repository) GetDB() *sql.DB {
	return u.DB
}

func New(ctx context.Context, db *sql.DB) *repository {
	return &repository{db}
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
