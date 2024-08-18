package sqlite

import "database/sql"

type Repository interface {
	GetDB() *sql.DB
}
