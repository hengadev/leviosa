package tests

import (
	"bytes"
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/api"
	"github.com/GaryHY/event-reservation-app/internal/database"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"testing"
	"time"
)

const (
	createUsersTable    = "CREATE TABLE IF NOT EXISTS users (email TEXT NOT NULL PRIMARY KEY, hashpassword TEXT NOT NULL);"
	createSessionsTable = "CREATE TABLE IF NOT EXISTS sessions (id TEXT NOT NULL PRIMARY KEY, email TEXT NOT NULL, created_at TEXT NOT NULL, expired_at TEXT NOT NULL);"
)

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func newGetRequest(id string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/event?id=%s", id), nil)
	return request
}

// TODO: Change that when the type Event is implemented
func newPostRequest(name string) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/event?name=%s", name), nil)
	return request
}

// Une fonction pour voir si les deux passwords correspondent
func assertPasswordHash(t testing.TB, got, want string) {
	t.Helper()
	if err := bcrypt.CompareHashAndPassword([]byte(got), []byte(want)); err != nil {
		t.Errorf("got the password %q from the database, expected %q", got, want)
		fmt.Println(err)
	}
}

// Une fonction pour check que j'ai bien qu'une seule fois le meme email dans la base de donnee
func assertEqualOne(t testing.TB, countEmail int, objectName string) {
	t.Helper()
	if countEmail != 1 {
		t.Errorf("got the count of %d, expected 1 %s", countEmail, objectName)
	}
}

func newPostJSONRequest(email, password, endpoint string) *http.Request {
	jsonData := []byte(fmt.Sprintf(`{"Email": "%s", "Password": "%s"}`, email, password))
	request, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return request
}

func assertCookieName(t testing.TB, got string) {
	t.Helper()
	if got != types.SessionCookieName {
		t.Errorf("Expected cookie's name to be %q, got %q", types.SessionCookieName, got)
	}
}

func assertIsUUID(t testing.TB, uuidString string) {
	t.Helper()
	if err := uuid.Validate(uuidString); err != nil {
		t.Errorf("Expected the cookie to have a valid uuid Value but got %q", uuidString)
	}
}

func assertSameExpirationDate(t testing.TB, got time.Time, want string) {
	t.Helper()
	timeParsedFromSession, _ := time.Parse(time.RFC822, want)
	timePlusExpirationDuration := timeParsedFromSession.Add(types.SessionDuration)
	expectedExpiration := timePlusExpirationDuration.Format(time.RFC822)

	gotExpiration := got.Local().Format(time.RFC822)
	if gotExpiration != expectedExpiration {
		t.Errorf("Expected cookie's %q, got expiration of %q", gotExpiration, expectedExpiration)
	}
}

func makeServerAndStoreWithUsersTable() (*api.Server, *sqlite.Store) {
	store, err := sqlite.NewStore("")
	if err != nil {
		log.Fatal("Something went wrong when creating the database")
	}
	store.Init(createUsersTable)
	server := api.NewServer(store)
	return server, store
}

func initUserTable(store *sqlite.Store) types.User {
	user := types.User{
		Email:    "test@example.fr",
		Password: "ThisisA_s@fe-pa22w0rd!",
	}
	if err := store.CreateUser(user); err != nil {
		log.Fatal("cannot create user in the test file - ", err)
	}
	return user
}
