package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestVerifyRejectsArguments(t *testing.T) {
	err := dispatch(context.Background(), []string{"verify", "extra"})
	if err == nil || !strings.Contains(err.Error(), "verify takes no arguments") {
		t.Fatalf("expected verify to reject arguments, got %v", err)
	}
}

func TestVerifyRequiresDatabaseURL(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production") // deterministic: no development.env loading
	t.Setenv("DATABASE_URL", "")

	err := dispatch(context.Background(), []string{"verify"})
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL is required") {
		t.Fatalf("expected verify to fail on a missing DATABASE_URL, got %v", err)
	}
	if code := exitCode(err); code != 1 {
		t.Fatalf("a configuration error must exit 1, got %d", code)
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"plain error", errors.New("boom"), 1},
		{"verify failure", failVerify(exitPepperMismatch, "mismatch"), exitPepperMismatch},
		{"wrapped verify failure", fmt.Errorf("outer: %w", failVerify(exitCanaryMissing, "missing")), exitCanaryMissing},
		{"non-positive status", failVerify(0, "zero"), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exitCode(tt.err); got != tt.want {
				t.Fatalf("exitCode = %d, want %d", got, tt.want)
			}
		})
	}
}

// The verify failures each have their own status, distinct from the generic 1.
func TestVerifyExitStatusesAreDistinct(t *testing.T) {
	seen := map[int]bool{1: true}
	for _, code := range []int{exitNothingToVerify, exitCanaryMissing, exitDecryptionFailed, exitPepperMismatch} {
		if seen[code] {
			t.Fatalf("exit status %d is used twice", code)
		}
		seen[code] = true
	}
}
