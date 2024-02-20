package tests

import (
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPOSTVote(t *testing.T) {
	server, store := makeServerAndStoreWithUsersTable()
	user := initUserTable(store)
	store.Init(createVotesTable, createSessionsTable, createEventsTable)

	tableTest := []struct {
		name          string
		doubleRequest bool
	}{
		{"Classic vote without complication expected", false},
		{"One user vote twice for the same event", true},
	}

	for _, tt := range tableTest {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanVotesTable(store)
			session := types.NewSession(user.Id)
			if err := store.CreateSession(session); err != nil {
				t.Error("failed to create the session - ", err)
			}

			cookie := &http.Cookie{
				Name:  types.SessionCookieName,
				Value: session.Id,
			}

			placecount_event := 20
			event := types.NewEvent(placecount_event)
			store.PostEvent(event)

			endpoint := fmt.Sprintf("/votes?id=%s", event.Id)
			request, _ := http.NewRequest(http.MethodPost, endpoint, nil)
			request.AddCookie(cookie)

			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			assertStatus(t, response.Code, http.StatusCreated)

			if tt.doubleRequest {
				newResponse := httptest.NewRecorder()
				server.ServeHTTP(newResponse, request)
				assertStatus(t, newResponse.Code, http.StatusBadRequest)
			}

			var lineCount int
			if err := store.DB.QueryRow("SELECT COUNT(*) FROM votes;").Scan(&lineCount); err != nil {
				t.Errorf("Cannot count the number of rows from the votes table")
			}
			assertEqualOne(t, lineCount, "vote")

			var got_userid string
			if err := store.DB.QueryRow("SELECT userid FROM votes;").Scan(&got_userid); err != nil {
				t.Errorf("Cannot find the userid from the votes table")
			}
			assertEqualString(t, got_userid, user.Id)

			var got_eventid string
			if err := store.DB.QueryRow("SELECT eventid FROM votes;").Scan(&got_eventid); err != nil {
				t.Errorf("Cannot find the eventid from the votes table")
			}
			assertEqualString(t, got_eventid, event.Id)

			var eventcount int
			if err := store.DB.QueryRow("SELECT placecount FROM events where id = ?;", event.Id).Scan(&eventcount); err != nil {
				t.Errorf("Cannot find the eventid from the votes table")
			}
			assertEqualInt(t, eventcount, placecount_event-1)
		})
	}

	noVoteTest := []struct {
		name             string
		postEventToStore bool
		statusCode       int
		placecount       int
	}{
		{"Event not in the database", false, http.StatusBadRequest, 20},
		{"User not authenticated", true, http.StatusForbidden, 20},
		{"No remaining place for the event", true, http.StatusBadRequest, 0},
	}

	for _, tt := range noVoteTest {
		t.Run(tt.name, func(t *testing.T) {
			placecount_event := tt.placecount
			event := types.NewEvent(placecount_event)
			if tt.postEventToStore {
				store.PostEvent(event)
			}

			endpoint := fmt.Sprintf("/votes?id=%s", event.Id)
			request, _ := http.NewRequest(http.MethodPost, endpoint, nil)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)
			assertStatus(t, response.Code, tt.statusCode)
		})
	}
}
