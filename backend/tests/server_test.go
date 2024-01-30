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
			store, err := sqlite.NewStubStore("")
			if err != nil {
				log.Fatal("Something went wrong when creating the database")
			}
			store.Init(createEventTable, insertValues)
			server := api.NewServer(store)

			request := newGetRequest(tt.id)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)
			assertStatus(t, response.Code, tt.eventStatus)
			assertResponseBody(t, response.Body.String(), tt.eventName)
		})
	}
}
