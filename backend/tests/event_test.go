package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/GaryHY/event-reservation-app/internal/api"
	"github.com/GaryHY/event-reservation-app/internal/database"
	"github.com/GaryHY/event-reservation-app/internal/types"
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

	eventTests := []struct {
		caseName    string
		eventStatus int
		event       types.Event
	}{
		{"Get name of event with ID 123", http.StatusOK, types.Event{Id: "1", Name: "event1"}},
		{"Get name of event with ID 456", http.StatusOK, types.Event{Id: "2", Name: "event2"}},
		{"Return 404 when event missing", http.StatusNotFound, types.Event{Id: "3", Name: ""}},
	}

	for _, tt := range eventTests {
		t.Run(tt.caseName, func(t *testing.T) {
			request := newGetRequest(tt.event.Id)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			var got types.Event
			err := json.NewDecoder(response.Body).Decode(&got)
			fmt.Printf("event from the store : [id: %s, name: %s]\n", got.Id, got.Name)
			if err != nil {
				t.Errorf("Unable to parse response from server %q into slice of Event, '%v'", response.Body, err)
			}

			assertStatus(t, response.Code, tt.eventStatus)
			if response.Code == http.StatusOK {
				if !reflect.DeepEqual(got, tt.event) {
					t.Errorf("got %v want %v", got, tt.event)
				}
			}
		})
	}
}

func TestPostEvent(t *testing.T) {
	// TODO: Add a test for when an event for a day is already planned, do that when the Event is fully implemented
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
