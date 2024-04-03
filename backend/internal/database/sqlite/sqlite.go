package sqlite

import (
	"database/sql"
	"fmt"
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
	store := &Store{db}
	// TODO: Make the queries for the db initialisation
	adminpassword := hashPassword("adminpassword")
	queries := []string{
		"CREATE TABLE IF NOT EXISTS users (id, email TEXT NOT NULL PRIMARY KEY, hashpassword TEXT NOT NULL, role TEXT NOT NULL, lastname TEXT NOT NULL, firstname TEXT NOT NULL, gender TEXT NOT NULL, birthdate TEXT NOT NULL, telephone TEXT NOT NULL, address TEXT NOTN NULL, city TEXT NOT NULL, postalcard INTEGER NOT NULL);",
		"CREATE TABLE IF NOT EXISTS sessions (id TEXT NOT NULL PRIMARY KEY, userid TEXT NOT NULL REFERENCES users, created_at TEXT NOT NULL);",
		"CREATE TABLE IF NOT EXISTS votes (id TEXT NOT NULL PRIMARY KEY, userid TEXT NOT NULL REFERENCES users, eventid TEXT NOT NULL REFERENCES events);",
		"CREATE TABLE IF NOT EXISTS events (id UUID NOT NULL PRIMARY KEY, location TEXT NOT NULL, placecount INTEGER NOT NULL, date TEXT NOT NULL, priceid TEXT NOT NULL);",
		fmt.Sprintf("INSERT INTO users VALUES (0, 'admin@example.fr', '%s', 'admin', 'adminlastname', 'adminfirstname', 'male', '20/08/1999', '0000 00 00 00', 'admin address', 'admin city', 'admin postalcard');", adminpassword),
	}
	for _, query := range queries {
		store.Init(query)
	}
	return store, nil
}

func (s *Store) Init(queries ...string) {
	for _, query := range queries {
		_, err := s.DB.Exec(query)
		if err != nil {
			log.Fatalf("Error executing '%s' : - %s", query, err)
		}
	}
}
