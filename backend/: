package tests

import (
	// "fmt"
	"github.com/GaryHY/event-reservation-app/internal/api"
	"github.com/GaryHY/event-reservation-app/internal/database"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGETEventByID(t *testing.T) {

	createEventTable := "CREATE TABLE IF NOT EXISTS events (id INTEGER NOT NULL PRIMARY KEY, name TEXT NOT NULL);"
	insertValues := "INSERT INTO events (name) VALUES ('event1'), ('event2');"

	store, err := sqlite.NewStore("")
	if err != nil {
		log.Fatal("Something went wrong when creating the database")
	}
	store.Init(createEventTable, insertValues)
	server := api.NewServer(store)

	eventTest := []struct {
		caseName    string
		id          int
		eventName   string
		eventStatus int
	}{
		{"Get name of event with ID 123", 1, "event1", http.StatusOK},
		{"Get name of event with ID 456", 2, "event2", http.StatusOK},
		{"Return 404 when event missing", 3, "", http.StatusNotFound},
	}

	for _, tt := range eventTest {
		t.Run(tt.caseName, func(t *testing.T) {

			request := newGetRequest(tt.id)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)
			assertStatus(t, response.Code, tt.eventStatus)
			assertResponseBody(t, response.Body.String(), tt.eventName)
		})
	}
}

func TestPostEvent(t *testing.T) {

	createEventTable := "CREATE TABLE IF NOT EXISTS events (id INTEGER NOT NULL PRIMARY KEY, name TEXT NOT NULL);"

	store, err := sqlite.NewStore("")
	if err != nil {
		log.Fatal("Something went wrong when creating the database")
	}
	store.Init(createEventTable)
	server := api.NewServer(store)
	eventName := "event1"

	request := newPostRequest(eventName)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)
	assertStatus(t, response.Code, http.StatusOK)

	var count int
	var name string
	countQuery := "SELECT COUNT(name) FROM events;"
	store.DB.QueryRow(countQuery).Scan(&count)
	if count != 1 {
		t.Errorf("got the count of %d, expected %d", count, 1)
	}

	nameQuery := "SELECT name FROM events;"
	store.DB.QueryRow(nameQuery).Scan(&name)
	if name != eventName {
		t.Errorf("got the name of %s, expected %s", name, eventName)
	}
}
