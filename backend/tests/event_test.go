package tests

import (
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

func TestGETUserEvents(t *testing.T) {
	server, store := makeServerAndStoreWithUsersTable()
	store.Init(createEventsTable, createSessionsTable, createVotesTable)
	user := initUserTable(store)
	session := types.NewSession(user.Id)

	cookie := &http.Cookie{
		Name:    types.SessionCookieName,
		Value:   session.Id,
		Expires: time.Now().Add(5 * time.Minute),
	}
	store.CreateSession(session)
	endpoint := fmt.Sprintf("/events?id=%s", user.Id)

	t.Run("Not authorized because not authenticated", func(t *testing.T) {
		userNotAuthenticatedId := uuid.NewString()
		endpoint := fmt.Sprintf("/events?id=%s", userNotAuthenticatedId)
		request, _ := http.NewRequest(http.MethodGet, endpoint, nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("No event in the database", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, endpoint, nil)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		want := []types.Event{} // no event to add to the database
		got := []types.Event{}
		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Errorf("Unable to parse response from server %q into slice of Events - '%v'", response.Body, err)
		}

		assertStatus(t, response.Code, http.StatusOK)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("With events in the database", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, endpoint, nil)
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

		for _, event := range want {
			_, err := store.DB.Exec("INSERT INTO events VALUES (?, ?, ?, ?, ?)", event.Id, event.Location, event.PlaceCount, event.Date.Format(time.RFC3339), event.PriceId)
			if err != nil {
				log.Fatal("Error executing the query insert in events table - ", err)
			}
			newVote := types.NewVote(&user.Id, &event.Id)
			_, err = store.DB.Exec("INSERT INTO votes VALUES (?, ?, ?)", newVote.Id, newVote.UserId, newVote.EventId)
			if err != nil {
				log.Fatal("Error executing the query insert in votes tables - ", err)
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
