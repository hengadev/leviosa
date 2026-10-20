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

func TestBootstrapAdminRejectsArguments(t *testing.T) {
	err := dispatch(context.Background(), []string{"bootstrap-admin", "extra"})
	if err == nil || !strings.Contains(err.Error(), "bootstrap-admin takes no arguments") {
		t.Fatalf("expected bootstrap-admin to reject arguments, got %v", err)
	}
}

func TestBootstrapAdminRequiresDatabaseURL(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production") // deterministic: no development.env loading
	t.Setenv("DATABASE_URL", "")

	err := dispatch(context.Background(), []string{"bootstrap-admin"})
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL is required") {
		t.Fatalf("expected bootstrap-admin to fail on a missing DATABASE_URL, got %v", err)
	}
}

func TestBootstrapAdminOutsideDevelopmentFailsOnMissingAdminValues(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("DATABASE_URL", "postgres://invalid.invalid/db") // present, never reached
	for _, key := range []string{
		"ADMIN_EMAIL", "ADMIN_PASSWORD", "ADMIN_FIRST_NAME", "ADMIN_LAST_NAME",
		"ADMIN_TELEPHONE", "ADMIN_POSTAL_CODE", "ADMIN_CITY", "ADMIN_ADDRESS1", "ADMIN_ADDRESS2",
	} {
		t.Setenv(key, "")
	}

	err := dispatch(context.Background(), []string{"bootstrap-admin"})
	if err == nil || !strings.Contains(err.Error(), "missing required values") {
		t.Fatalf("expected bootstrap-admin to fail on missing ADMIN_* values outside development, got %v", err)
	}
}

func TestMigrateRejectsArguments(t *testing.T) {
	err := dispatch(context.Background(), []string{"migrate", "extra"})
	if err == nil || !strings.Contains(err.Error(), "migrate takes no arguments") {
		t.Fatalf("expected migrate to reject arguments, got %v", err)
	}
}
