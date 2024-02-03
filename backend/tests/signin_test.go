package tests

import (
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPOSTSignIn(t *testing.T) {
	// TODO: test a faire
	// 1. aucun cookie set, l'utilisateur est dans la base de donnee
	// 2. L'utilisateur n'est pas dans la base de donnee, il n'est pas  register.
	// 3. si la personne est deja register donc si un cookie est deja present
	// 4. il y a un cookie mais le cookies est expire

	SignInWithCookieTests := []struct {
		name         string
		httpStatus   int
		hasCookieSet bool
	}{
		{"User logs without any active session", http.StatusOK, false},
		{"Has a session registered with cookie non expired", http.StatusOK, true},
		// NOTE: Cas special ou l'utilisateur n'est pas dans la bases de donnee mais un cookie ?
		// {"User not in database", "test-not-in-database@example.fr", "123", http.StatusConflict, true, true, true},
		// {"User not in database", "test-not-in-database@example.fr", "123", http.StatusConflict, true, false, true},
		// {"User already registered with no expired cookie", "test-not-in-database@example.fr", "123", http.StatusConflict, true, false, true},
		// {"User in database with expired cookie", "test-not-in-database@example.fr", "123", http.StatusConflict, true, false, true},
	}
	server, store := makeServerAndStoreWithUsersTable()
	store.Init(createSessionsTable)
	user := initUserTable(store)

	for _, tt := range SignInWithCookieTests {
		t.Run(tt.name, func(t *testing.T) {
			// NOTE: On fait une premiere requete pour set la session puis on check avec une nouvelle requete que je ne cree pas de nouvelle session
			request := newPostJSONRequest(user.Email, user.Password, "/signin")
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			assertStatus(t, response.Code, http.StatusOK)
			cookies := response.Result().Cookies()
			if tt.hasCookieSet {
				server.ServeHTTP(response, request)
				assertStatus(t, response.Code, http.StatusOK)
			}
			session := types.NewSession(user)
			assertEqualOne(t, len(cookies), "cookie")
			assertCookieName(t, cookies[0].Name)
			assertIsUUID(t, cookies[0].Value)
			assertSameExpirationDate(t, cookies[0].Expires, session.Created_at)

			var lineCount int
			store.DB.QueryRow("SELECT COUNT(*) FROM sessions;").Scan(&lineCount)
			assertEqualOne(t, lineCount, "session saved")
			var email string
			store.DB.QueryRow("SELECT email FROM sessions;").Scan(&email)
			if email != user.Email {
				t.Errorf("wrong mail registered in database : got %q, want %q", email, user.Email)
			}
			defer store.DeleteSession(session)
		})
	}

	WrongIdTests := []struct {
		name       string
		email      string
		password   string
		statusCode int
	}{
		{"Wrong password", user.Email, "123wrongpassword", http.StatusUnauthorized},
		{"Wrong email", "wrongmail@example.fr", user.Password, http.StatusUnauthorized},
	}
	for _, tt := range WrongIdTests {
		t.Run(tt.name, func(t *testing.T) {
			request := newPostJSONRequest(tt.email, tt.password, "/signin")
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)
			assertStatus(t, response.Code, tt.statusCode)
		})
	}
}
