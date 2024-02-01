package tests

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"testing"
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
	}
}

// Une fonction pour check que j'ai bien qu'une seule fois le meme email dans la base de donnee
func assertEmailCount(t testing.TB, countEmail int) {
	if countEmail != 1 {
		t.Errorf("got the count of %d, expected 1", countEmail)
	}
}
