package tests

import (
	"bytes"
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/api"
	"github.com/GaryHY/event-reservation-app/internal/database"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/google/uuid"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TODO: Test the validate email/password functions

func TestPOSTSignUp(t *testing.T) {
	SignUpTests := []struct {
		name        string
		email       string
		password    string
		httpStatus  int // TODO: Check the  status I should send for each case
		succesQuery bool
	}{
		{"Valid email and valid password", "henry.gary@hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusCreated, true},
		{"User already in database", "henry.gary@hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusConflict, true},
		{"Invalid email and valid password", "henry.gary/hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusForbidden, false},
		{"Valid email and invalid password", "henry.gary/hotmail.com", "tHi#sisa/GoOd-password_12", http.StatusForbidden, false},
	}

	// TODO: Put that in a function create Server
	createUsersTable := "CREATE TABLE IF NOT EXISTS users (email TEXT NOT NULL PRIMARY KEY, hashpassword TEXT NOT NULL);"

	store, err := sqlite.NewStore("")
	if err != nil {
		log.Fatal("Something went wrong when creating the database")
	}
	store.Init(createUsersTable)
	server := api.NewServer(store)

	for _, tt := range SignUpTests {
		t.Run(tt.name, func(t *testing.T) {

			jsonData := []byte(fmt.Sprintf(`{"Email": "%s", "Password": "%s"}`, tt.email, tt.password))
			request, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonData))
			request.Header.Set("Content-Type", "application/json; charset=UTF-8")
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)
			assertStatus(t, response.Code, tt.httpStatus)

			if tt.succesQuery {
				var hashpassword string
				var countEmail int

				store.DB.QueryRow("SELECT COUNT(email) FROM users;").Scan(&countEmail)
				assertEqualOne(t, countEmail, "email")
				store.DB.QueryRow("SELECT hashpassword FROM users;").Scan(&hashpassword)
				assertPasswordHash(t, hashpassword, tt.password)
			}
		})
	}

}

func TestPOSTSignIn(t *testing.T) {
	// TODO: test a faire
	// 1. aucun cookie set, l'utilisateur est dans la base de donnee
	// 2. L'utilisateur n'est pas dans la base de donnee, il n'est pas  register.
	// 3. si la personne est deja register donc si un cookie est deja present
	// 4. il y a un cookie mais le cookies est expire

	SignInTests := []struct {
		name            string
		email           string
		password        string
		httpStatus      int // TODO: Check the  status I should send for each case
		gotCookie       bool
		isCookieExpired bool
		expectCookie    bool
	}{
		{"User exists in database", "test@example.fr", "123", http.StatusOK, false, false, true},
		// NOTE: Cas special ou l'utilisateur n'est pas dans la bases de donnee mais un cookie ?
		// {"User not in database", "test-not-in-database@example.fr", "123", http.StatusConflict, true, true, true},
		// {"User not in database", "test-not-in-database@example.fr", "123", http.StatusConflict, true, false, true},

		// {"User already registered with no expired cookie", "test-not-in-database@example.fr", "123", http.StatusConflict, true, false, true},
		// {"User in database with expired cookie", "test-not-in-database@example.fr", "123", http.StatusConflict, true, false, true},
		// {"Wrong password", "test@example.fr", "123wrongpassword", http.StatusConflict, true},
	}
	_ = SignInTests

	createUsersTable := "CREATE TABLE IF NOT EXISTS users (email TEXT NOT NULL PRIMARY KEY, hashpassword TEXT NOT NULL);"
	createSessionsTable := "CREATE TABLE IF NOT EXISTS sessions (id TEXT NOT NULL PRIMARY KEY, email TEXT NOT NULL, created_at TEXT NOT NULL, expired_at TEXT NOT NULL);"
	store, err := sqlite.NewStore("")
	if err != nil {
		log.Fatal("Something went wrong when creating the database")
	}
	store.Init(createUsersTable, createSessionsTable)
	server := api.NewServer(store)

	t.Run("User exists in database", func(t *testing.T) {
		user := types.User{
			Email:    "test@example.fr",
			Password: "ThisisA_s@fe-pa22w0rd!",
		}
		if err := store.CreateUser(user); err != nil {
			log.Fatal("cannot create user in the test file - ", err)
		}

		jsonData := []byte(fmt.Sprintf(`{"Email": "%s", "Password": "%s"}`, user.Email, user.Password))
		request, _ := http.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(jsonData))
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
		session := types.NewSession(user)
		cookies := response.Result().Cookies()
		assertEqualOne(t, len(cookies), "cookie")

		if cookies[0].Name != types.SessionCookieName {
			t.Errorf("Expected cookie's name to be %q, got %q", "session_token", cookies[0].Name)
		}

		if err := uuid.Validate(cookies[0].Value); err != nil {
			t.Errorf("Expected the cookie to have a valid uuid Value but got %q", cookies[0].Value)
		}

		creationFromSession, _ := time.Parse(time.RFC822, session.Created_at)
		creationAndExpiry := creationFromSession.Add(types.SessionDuration)
		sessionFormat := creationAndExpiry.Format(time.RFC822)

		cookieBase := cookies[0].Expires.Local().Format(time.RFC822)
		if cookieBase != sessionFormat {
			t.Errorf("Expected cookie's %q, got expiration of %q", cookieBase, sessionFormat)
		}

		var lineCount int
		store.DB.QueryRow("SELECT COUNT(*) FROM sessions;").Scan(&lineCount)
		assertEqualOne(t, lineCount, "session saved")

		var email string
		store.DB.QueryRow("SELECT email FROM sessions;").Scan(&email)
		if email != user.Email {
			t.Errorf("wrong mail registered in database : got %q, want %q", email, user.Email)
		}
		// TODO: Check the expire from the database when done with what is above, you've got a lot to do

	})
}
