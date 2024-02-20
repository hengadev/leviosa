package tests

import (
	// "github.com/GaryHY/event-reservation-app/internal/api"
	// "github.com/GaryHY/event-reservation-app/internal/database"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/google/uuid"
	"net/http"
	"time"
	// "fmt"
	"log"
	"net/http/httptest"
	"testing"
)

func TestPOSTSignOut(t *testing.T) {
	server, store := makeServerAndStoreWithUsersTable()
	user := initUserTable(store)
	store.Init(createSessionsTable, createUsersTable)

	t.Run("User do not have active session", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodPost, "/signout", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("User has an active session", func(t *testing.T) {
		uuid := uuid.NewString()
		cookie := &http.Cookie{
			Name:  types.SessionCookieName,
			Value: uuid,
		}

		request, _ := http.NewRequest(http.MethodPost, "/signout", nil)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()

		// session := types.Session{Id: uuid, Email: user.Email, Created_at: time.Now().Add(5 * time.Minute)}
		session := types.Session{Id: uuid, UserId: user.Email, Created_at: time.Now().Add(5 * time.Minute)}
		store.CreateSession(&session)

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		var count int
		if err := store.DB.QueryRow("SELECT COUNT(*) FROM sessions;").Scan(&count); err != nil {
			log.Fatal("Cannot count the sessions from the sessions table")
		}
		if count != 0 {
			t.Errorf("Got count of %d, expected %d", count, 0)
		}
	})
}
