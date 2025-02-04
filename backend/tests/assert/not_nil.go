package assert

import "testing"

func NotNil(t testing.TB, got error) {
	t.Helper()
	if got != nil {
		t.Errorf("got %q, want nil error", got)
	}
}
