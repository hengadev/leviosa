package session

import (
	"testing"
)

func TestCreateSession(t *testing.T) {
	// TODO: Write all the tests that I need.
	tests := []Test{
		{name: "the user already has a valid sessionID"},
		{name: "the user already has a non valid sessionID"},
	}
	_ = tests
}

type Test struct {
	name string
}
