package assert

import (
	"reflect"
	"testing"
)

func ReflectNotEqual[T any](t testing.TB, got, want T) {
	t.Helper()
	if reflect.DeepEqual(got, want) {
		t.Errorf("got '%v', want '%v'", got, want)
	}
}
