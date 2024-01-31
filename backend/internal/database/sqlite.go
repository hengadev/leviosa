package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type Store struct {
	DB *sql.DB
}

func NewStore(connStr string) (*Store, error) {
	if connStr == "" {
		connStr = ":memory:"
	}
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log.Fatal("Cannot connect to the database - ", err)
	}
	return &Store{db}, nil
}

func (s *Store) Init(queries ...string) {
	for _, query := range queries {
		_, err := s.DB.Exec(query)
		if err != nil {
			log.Fatalf("Error executing '%s' : - %s", query, err)
		}
	}
}

func (s *Store) GetEventNameByID(id string) (event types.Event) {
	query := fmt.Sprintf("SELECT * FROM events WHERE id=%s;", id)
	s.DB.QueryRow(query).Scan(&event.Id, &event.Name)
	return
}

func (s *Store) PostEvent(name string) {
	query := fmt.Sprintf("INSERT INTO events (name) VALUES ('%s');", name)
	_, err := s.DB.Exec(query)
	if err != nil {
		log.Fatal("Could not insert into the database - ", err)
	}
}
