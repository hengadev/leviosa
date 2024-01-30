package sqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type StubStore struct {
	db *sql.DB
}

func NewStubStore(connStr string) (*StubStore, error) {
	if connStr == "" {
		connStr = ":memory:"
	}
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log.Fatal("Cannot connect to the database - ", err)
	}
	return &StubStore{db}, nil
}

func (s *StubStore) Init(queries ...string) {
	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			log.Fatalf("Error executing '%s' : - %s", query, err)
		}
	}
}

func (s *StubStore) GetEventNameByID(id string) (name string) {
	query := fmt.Sprintf("SELECT name FROM events WHERE id=%s;", id)
	s.db.QueryRow(query).Scan(&name)
	return name
}
