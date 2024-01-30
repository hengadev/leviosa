package tests

import (
	"fmt"
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

func newGetRequest(id int) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/event/%d", id), nil)
	return request
}
