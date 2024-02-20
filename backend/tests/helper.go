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
	createUsersTable    = "CREATE TABLE IF NOT EXISTS users (id, email TEXT NOT NULL PRIMARY KEY, hashpassword TEXT NOT NULL, role TEXT NOT NULL, lastname TEXT NOT NULL, firstname TEXT NOT NULL, gender TEXT NOT NULL, birthdate TEXT NOT NULL, telephone TEXT NOT NULL, address TEXT NOTN NULL, city TEXT NOT NULL, postalcard INTEGER NOT NULL);"
	createSessionsTable = "CREATE TABLE IF NOT EXISTS sessions (id TEXT NOT NULL PRIMARY KEY, userid TEXT NOT NULL REFERENCES users, created_at TEXT NOT NULL);"
	createVotesTable    = "CREATE TABLE IF NOT EXISTS votes (id TEXT NOT NULL PRIMARY KEY, userid TEXT NOT NULL REFERENCES users, eventid TEXT NOT NULL REFERENCES events);"
	createEventsTable   = "CREATE TABLE IF NOT EXISTS events (id UUID NOT NULL PRIMARY KEY, location TEXT NOT NULL, placecount INTEGER NOT NULL, date TEXT NOT NULL);"
)

func assertEqualString(t testing.TB, got, want string) {
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

// Une fonction pour voir si les deux passwords correspondent
func assertPasswordHash(t testing.TB, got, want string) {
	t.Helper()
	if err := bcrypt.CompareHashAndPassword([]byte(got), []byte(want)); err != nil {
		t.Errorf("got the password %q from the database, expected %q", got, want)
		fmt.Println(err)
	}
}

// Une fonction pour check que j'ai bien qu'une seule fois le meme email dans la base de donnee
func assertEqualOne(t testing.TB, objectCount int, objectName string) {
	t.Helper()
	if objectCount != 1 {
		t.Errorf("got the count of %d, expected 1 %s", objectCount, objectName)
	}
}

func assertEqualInt(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, expected %d", got, want)
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

// TODO: Remove that function since it will be stored using the time.Time function
func assertSameExpirationDate(t testing.TB, got, want time.Time) {
	t.Helper()
	gotParsed := got.Local().Format(time.RFC822)
	wantParsed := want.Add(types.SessionDuration).Format(time.RFC822)

	if gotParsed != wantParsed {
		// t.Errorf("Expected cookie's %q, got expiration of %q", gotExpiration, expectedExpiration)
		t.Errorf("Expected cookie's %q, got expiration of %q", wantParsed, gotParsed)
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

// func initUserTable(store *sqlite.Store) *types.UserStored {
func initUserTable(store *sqlite.Store) *types.UserStored {
	userForm := &types.UserForm{
		Email:      "test@example.fr",
		Password:   "ThisisA_s@fe-pa22w0rd!",
		Role:       string(types.BASIC),
		LastName:   "",
		FirstName:  "",
		Gender:     "",
		BirthDate:  "",
		Telephone:  "",
		Address:    "",
		City:       "",
		PostalCard: "",
	}
	user := types.NewUserStored(userForm)
	if err := store.CreateUser(user, false); err != nil {
		log.Fatal("cannot create user in the test file - ", err)
	}
	return user
}

func cleanSessionTable(store *sqlite.Store) {
	_, err := store.DB.Exec("DELETE FROM sessions;")
	if err != nil {
		log.Fatal("Cannot delete everything from the user table", err)
	}
}

func cleanVotesTable(store *sqlite.Store) {
	_, err := store.DB.Exec("DELETE FROM votes;")
	if err != nil {
		log.Fatal("Cannot delete everything from the user table", err)
	}
}
