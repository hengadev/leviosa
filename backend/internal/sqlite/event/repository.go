package eventRepository

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type EventRepository struct {
	DB *sql.DB
}

func (e *EventRepository) GetDB() *sql.DB {
	return e.DB
}

func New(ctx context.Context, db *sql.DB) *EventRepository {
	return &EventRepository{db}
}
