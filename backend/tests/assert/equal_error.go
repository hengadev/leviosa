package assert

import (
	"errors"
	"testing"
)

func EqualError(t testing.TB, got, want error) {
	switch {
	case got == nil && want == nil:
		return
	case got == nil || want == nil:
		t.Errorf("got '%v', want '%v'", got, want)
	case !errors.Is(got, want):
		t.Errorf("got '%v', want '%v'", got.Error(), want.Error())
	}
}
