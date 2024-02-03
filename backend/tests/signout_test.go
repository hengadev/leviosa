package tests

import (
	// "github.com/GaryHY/event-reservation-app/internal/api"
	// "github.com/GaryHY/event-reservation-app/internal/database"
	// "github.com/GaryHY/event-reservation-app/internal/types"
	// "net/http"
	// "fmt"
	"net/http/httptest"
	"testing"
)

func TestPOSTSignOut(t *testing.T) {
	// TODO: test a faire
	// 2. user not in database
	// 1. user in database and decide to sign out

	server, store := makeServerAndStoreWithUsersTable()

	user := initUserTable(store)
	store.Init(createSessionsTable)

	t.Run("User not in database", func(t *testing.T) {
		request := newPostJSONRequest(user.Email, user.Password, "/signin")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		// TODO: Create session for the user
		// session := types.NewSession(user)
		// cookies := response.Result().Cookies()

		request = newPostJSONRequest(user.Email, user.Password, "/signout")
		response = httptest.NewRecorder()
		//
		// server.ServeHTTP(response, request)
		// assertStatus(t, response.Code, http.StatusOK)
	})
}
