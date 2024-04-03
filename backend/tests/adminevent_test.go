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
	server, store := makeServerAndStoreWithUsersTable()
	store.Init(createEventsTable, createSessionsTable)
	user := initUserTableAdmin(store)
	session := types.NewSession(user.Id)

	cookie := &http.Cookie{
		Name:    types.SessionCookieName,
		Value:   session.Id,
		Expires: time.Now().Add(5 * time.Minute),
	}
	store.CreateSession(session)

	t.Run("Not authorized because not admin", func(t *testing.T) {
		// NOTE: The http.MethodGet here is random, any method would give the same result
		request, _ := http.NewRequest(http.MethodGet, "/admin/events", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("No event in the database", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/admin/events", nil)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		want := []types.Event{}
		got := []types.Event{}
		if err = json.NewDecoder(response.Body).Decode(&got); err != nil {
			t.Errorf("Unable to parse response from server %q into slice of Event - '%v'", response.Body, err)
		}

		assertStatus(t, response.Code, http.StatusNotFound)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("With events in the database", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, "/admin/events", nil)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()

		time1, err := time.Parse(time.RFC3339, "2024-04-02T13:39:43.000Z")
		if err != nil {
			log.Fatal("Failed parse in Location time - ", err)
		}

		want := []types.Event{
			{Id: uuid.NewString(), Location: "Somewhere", PlaceCount: 40, Date: time1, PriceId: ""},
			{Id: uuid.NewString(), Location: "Some other place", PlaceCount: 32, Date: time1, PriceId: ""},
		}

		//put the want in the database
		for _, event := range want {
			_, err := store.DB.Exec("INSERT INTO events (id, location, placecount, date, priceid) VALUES (?, ?, ?, ?, ?)", event.Id, event.Location, event.PlaceCount, time1.Format(time.RFC3339), event.PriceId)
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
	server, store := makeServerAndStoreWithUsersTable()
	store.Init(createEventsTable, createSessionsTable)
	user := initUserTableAdmin(store)
	session := types.NewSession(user.Id)

	cookie := &http.Cookie{
		Name:    types.SessionCookieName,
		Value:   session.Id,
		Expires: time.Now().Add(5 * time.Minute),
	}
	store.CreateSession(session)

	// NOTE: The formatting for the time sent should be time.RFC3339

	time1, err := time.Parse(time.RFC3339, "2024-04-02T13:39:43.000Z")

	if err != nil {
		log.Fatal("Failed parse time - ", err)
	}

	want := &types.EventForm{
		Location:   "Somewhere",
		PlaceCount: 28,
		Date:       time1,
	}

	jsonData := []byte(fmt.Sprintf(`{"location": "%s", "placecount": %d, "date": "%s"}`, want.Location, want.PlaceCount, want.Date.Format(time.RFC3339)))
	request, _ := http.NewRequest(http.MethodPost, "/admin/events", bytes.NewBuffer(jsonData))
	request.AddCookie(cookie)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusCreated)
	// NOTE: check if that function return the time with time.RFC3339 formatting
	got := store.GetAllEvents()[0]

	assertIsUUID(t, got.Id)
	assertEqualString(t, got.Location, want.Location)
	assertEqualInt(t, got.PlaceCount, want.PlaceCount)
	if got.Date != want.Date {
		t.Errorf("\ngot %v\n want %v", got.Date, want.Date)
	}
	// TODO: add test to see if the product is created in stripe.
}

func TestDELETEEvent(t *testing.T) {
	// TODO: add test for event not in database
	server, store := makeServerAndStoreWithUsersTable()
	store.Init(createEventsTable, createSessionsTable)

	user := initUserTableAdmin(store)
	session := types.NewSession(user.Id)

	cookie := &http.Cookie{
		Name:    types.SessionCookieName,
		Value:   session.Id,
		Expires: time.Now().Add(5 * time.Minute),
	}
	store.CreateSession(session)

	tableTest := []struct {
		name             string
		postEventToStore bool
		statusCode       int
	}{
		{"delete event that is in database", true, http.StatusNoContent},
		{"event is not in database", false, http.StatusBadRequest},
	}

	for _, tt := range tableTest {
		t.Run(tt.name, func(t *testing.T) {
			event := types.NewEvent("Some Location", 40, time.Now(), "")
			if tt.postEventToStore {
				store.PostEvent(event)
			}

			endpoint := fmt.Sprintf("/admin/events?id=%s", event.Id)
			request, _ := http.NewRequest(http.MethodDelete, endpoint, nil)
			request.AddCookie(cookie)

			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			assertStatus(t, response.Code, tt.statusCode)
			var countEvent int
			store.DB.QueryRow("SELECT count(*) FROM events WHERE id=?", event.Id).Scan(&countEvent)
			assertEqualInt(t, countEvent, 0)
		})
	}
}

func TestUPDATEEvent(t *testing.T) {
	// TODO: add test for event not in database
	server, store := makeServerAndStoreWithUsersTable()
	store.Init(createEventsTable, createSessionsTable)

	user := initUserTableAdmin(store)
	session := types.NewSession(user.Id)

	cookie := &http.Cookie{
		Name:    types.SessionCookieName,
		Value:   session.Id,
		Expires: time.Now().Add(5 * time.Minute),
	}
	store.CreateSession(session)

	tableTest := []struct {
		name             string
		postEventToStore bool
		statusCode       int
	}{
		{"event is not in database", false, http.StatusBadRequest},
		{"update event that is in database", true, http.StatusCreated},
	}

	for _, tt := range tableTest {
		t.Run(tt.name, func(t *testing.T) {
			event := types.NewEvent("Some Location", 40, time.Now(), "")
			if tt.postEventToStore {
				store.PostEvent(event)
			}

			newLocation := "My new location"
			newPlacecount := 120
			newDate := time.Now().Add(5 * time.Hour)
			newDateStr := newDate.Format(time.RFC3339)
			jsonData := []byte(fmt.Sprintf(`{"id": "%s", "location": "%s", "placecount": %d, "date": "%s", "priceid": ""}`, event.Id, newLocation, newPlacecount, newDateStr))

			endpoint := fmt.Sprintf("/admin/events?id=%s", event.Id)
			request, _ := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer([]byte(jsonData)))
			request.AddCookie(cookie)

			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			assertStatus(t, response.Code, tt.statusCode)

			var countEvent int
			store.DB.QueryRow("SELECT COUNT(*) FROM events WHERE id=?", event.Id).Scan(&countEvent)

			if tt.postEventToStore {
				var placecount int
				store.DB.QueryRow("SELECT placecount FROM events WHERE id=?", event.Id).Scan(&placecount)
				assertEqualInt(t, placecount, newPlacecount)

				var location string
				store.DB.QueryRow("SELECT location FROM events WHERE id=?", event.Id).Scan(&location)
				assertEqualString(t, location, newLocation)

				var date string
				store.DB.QueryRow("SELECT date FROM events WHERE id=?", event.Id).Scan(&date)
				// assertEqualString(t, date, newDate.Format(time.RFC822))
				assertEqualString(t, date, newDateStr)

				assertEqualInt(t, countEvent, 1)
			} else {
				assertEqualInt(t, countEvent, 0)
			}
		})
	}
}
