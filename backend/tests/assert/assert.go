package assert

import (
	"testing"
)

// Here do a bunch of things to just test all the assertions that I have in my tests.
// TODO: do most of the function that you get on testify and be sure to use generic when necessary

// the generic way to compare two values of the same type
func Equal[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func NotEqual[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got == want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func NotNil(t testing.TB, got error) {
	t.Helper()
	if got != nil {
		t.Errorf("got %q, want nil error", got)
	}
}
