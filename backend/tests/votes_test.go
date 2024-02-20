package tests

import (
	// "bytes"
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPOSTVote(t *testing.T) {
	// FIX:
	// 1. vote classic
	// 2. user not auth, should return http.StatusForbidden
	// NOTE: test a faire
	// 3. event (send in the body) does not exist, should return http.StatusBadRequest
	// 4. already voted for that event
	// 5. vote for an event that you can not vote for because it has 0 place left - should return an http.StatusBadRequest because that event should not have been displayed in the first place

	server, store := makeServerAndStoreWithUsersTable()
	user := initUserTable(store)
	store.Init(createVotesTable, createSessionsTable, createEventsTable)

	t.Run("Classic vote without complication expected", func(t *testing.T) {
		session := types.NewSession(&types.User{Email: user.Email, Password: user.Password}, user.Id)
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
		// NOTE: Something when I want to make the query when the user has already made one
		// if something {
		// 	server.ServeHTTP(response, request)
		// }

		assertStatus(t, response.Code, http.StatusCreated)

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
