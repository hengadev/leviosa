package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/google/uuid"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestGETEvents(t *testing.T) {

	// TODO: Complete the create event table with all the fields when the schema is done
	createEventTable := "CREATE TABLE IF NOT EXISTS events (id UUID NOT NULL PRIMARY KEY, location TEXT NOT NULL, placecount INTEGER NOT NULL, date TEXT NOT NULL);"

	server, store := makeServerAndStoreWithUsersTable()
	store.Init(createEventTable)

	t.Run("No event in the database", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/event", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		want := []types.Event{}
		got := []types.Event{}
		err = json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Errorf("Unable to parse response from server %q into slice of Event, '%v'", response.Body, err)
		}
		assertStatus(t, response.Code, http.StatusNotFound)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("With events in the database", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/event", nil)
		response := httptest.NewRecorder()

		time1, err := time.Parse(types.EventFormat, "2024-02-13")
		if err != nil {
			log.Fatal("Failed parse in Location time - ", err)
		}
		time2 := time1.Format(types.EventFormat)

		want := []types.Event{
			types.Event{Id: uuid.NewString(), Location: "Somewhere", PlaceCount: 40, Date: time2},
			types.Event{Id: uuid.NewString(), Location: "Some other place", PlaceCount: 32, Date: time2},
		}

		for _, event := range want {
			_, err := store.DB.Exec("INSERT INTO events (id, location, placecount, date) VALUES (?, ?, ?, ?)", event.Id, event.Location, event.PlaceCount, event.Date)
			if err != nil {
				log.Fatal("Error executing the query insert in database - ", err)
			}
		}

		server.ServeHTTP(response, request)

		got := []types.Event{}
		if err = json.NewDecoder(response.Body).Decode(&got); err != nil {
			t.Errorf("Unable to parse response from server %q into slice of Event, '%v'", response.Body, err)
		}

		assertStatus(t, response.Code, http.StatusOK)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("\ngot %v\n want %v", got, want)
		}
	})

}

func TestPOSTEvent(t *testing.T) {
	// TODO: Add a test for when an event for a day is already planned, do that when the Event is fully implemented

	createEventTable := "CREATE TABLE IF NOT EXISTS events (id UUID NOT NULL PRIMARY KEY, location TEXT NOT NULL, placecount INTEGER NOT NULL, date TEXT NOT NULL);"

	server, store := makeServerAndStoreWithUsersTable()
	store.Init(createEventTable)

	time1, err := time.Parse(types.EventFormat, "2024-02-13")
	if err != nil {
		log.Fatal("Failed parse time - ", err)
	}
	time2 := time1.Format(types.EventFormat)

	want := types.Event{
		Id:         uuid.NewString(),
		Location:   "Somewhere",
		PlaceCount: 28,
		Date:       time2,
	}

	jsonData := []byte(fmt.Sprintf(`{"id": "%s", "location": "%s", "placeCount": %d, "date": "%s"}`, want.Id, want.Location, want.PlaceCount, want.Date))
	request, _ := http.NewRequest(http.MethodPost, "/event", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusOK)
	got := store.GetAllEvents()[0]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot %v\n want %v", got, want)
	}
}
