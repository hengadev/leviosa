package main

import (
	"context"
	"strings"
	"testing"
)

func TestDispatchUnknownModeListsAvailableModes(t *testing.T) {
	err := dispatch(context.Background(), []string{"nope"})
	if err == nil {
		t.Fatal("expected an error for an unknown mode")
	}
	if !strings.Contains(err.Error(), `unknown mode "nope"`) || !strings.Contains(err.Error(), "serve") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServeRejectsArguments(t *testing.T) {
	err := dispatch(context.Background(), []string{"serve", "extra"})
	if err == nil || !strings.Contains(err.Error(), "serve takes no arguments") {
		t.Fatalf("expected serve to reject arguments, got %v", err)
	}
}
