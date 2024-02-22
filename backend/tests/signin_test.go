package tests

import (
	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPOSTSignIn(t *testing.T) {
	server, store := makeServerAndStoreWithUsersTable()
	store.Init(createSessionsTable)
	user := initUserTable(store)

	wrongGetRequest, _ := http.NewRequest(http.MethodGet, "/signin", nil)

	testWithoutCookie := []struct {
		name       string
		request    *http.Request
		statusCode int
	}{
		{"User email or password in the wrong format", newPostJSONRequest("abc", user.Password, "/signin"), http.StatusForbidden},
		{"User is not in the database", newPostJSONRequest("userNotinDatabase@example.fr", user.Password, "/signin"), http.StatusUnauthorized},
		{"Wrong method used", wrongGetRequest, http.StatusMethodNotAllowed},
	}

	for _, tt := range testWithoutCookie {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			server.ServeHTTP(response, tt.request)
			assertStatus(t, response.Code, tt.statusCode)
		})
	}

	t.Run("Basic sign in with everything according to plan", func(t *testing.T) {
		request := newPostJSONRequest(user.Email, user.Password, "/signin")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
		cookies := response.Result().Cookies()
		session := types.NewSession(user.Id)
		assertEqualOne(t, len(cookies), "cookie")
		assertCookieName(t, cookies[0].Name)
		assertIsUUID(t, cookies[0].Value)
		assertSameExpirationDate(t, cookies[0].Expires, session.Created_at)

		var lineCount int
		store.DB.QueryRow("SELECT COUNT(*) FROM sessions;").Scan(&lineCount)
		assertEqualOne(t, lineCount, "sessions saved")

		var userCount int
		store.DB.QueryRow("SELECT COUNT(*) FROM users;").Scan(&userCount)
		assertEqualOne(t, userCount, "users saved")

		var id string
		store.DB.QueryRow("SELECT userid FROM sessions;").Scan(&id)
		if id != user.Id {
			t.Errorf("wrong user_id registered in database : got %q, want %q", id, user.Id)
		}
		defer cleanSessionTable(store)
	})

	testWithCookie := []struct {
		name               string
		session_created_at time.Time
		new_cookie         bool
	}{
		{"User already registered with valid cookie", time.Now(), false},
		{"User has an expired (non valid) cookie", time.Now().Add(-types.SessionDuration), true},
	}

	for _, tt := range testWithCookie {
		t.Run(tt.name, func(t *testing.T) {
			request := newPostJSONRequest(user.Email, user.Password, "/signin")
			uuid := uuid.NewString()
			cookie := &http.Cookie{
				Name:  types.SessionCookieName,
				Value: uuid,
			}
			request.AddCookie(cookie)
			// session := types.Session{Id: uuid, Email: user.Email, Created_at: tt.session_created_at}
			session := types.Session{Id: uuid, UserId: user.Email, Created_at: tt.session_created_at}
			store.CreateSession(&session)

			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			assertStatus(t, response.Code, http.StatusOK)

			var lineCount int
			store.DB.QueryRow("SELECT COUNT(*) FROM sessions;").Scan(&lineCount)
			assertEqualOne(t, lineCount, "session saved")
			var db_id string
			store.DB.QueryRow("SELECT id FROM sessions;").Scan(&db_id)
			if tt.new_cookie {
				if db_id == uuid {
					t.Errorf("wrong ID registered in database : got %q, want %q", db_id, uuid)
				}
			} else {
				if db_id != uuid {
					t.Errorf("wrong ID registered in database : got %q, want %q", db_id, uuid)
				}
			}
			defer cleanSessionTable(store)
		})
	}
}
