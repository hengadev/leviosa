package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthReportsStatusAndDefaultSHA(t *testing.T) {
	rec := httptest.NewRecorder()
	handleHealth(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Status string `json:"status"`
		SHA    string `json:"sha"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("expected status %q, got %q", "ok", body.Status)
	}
	if body.SHA != "dev" {
		t.Errorf("expected default sha %q, got %q", "dev", body.SHA)
	}
}

func TestHealthReportsInjectedSHA(t *testing.T) {
	// Simulate the -ldflags -X injection used at build time.
	previous := ReleaseSHA
	ReleaseSHA = "0123456789abcdef0123456789abcdef01234567"
	defer func() { ReleaseSHA = previous }()

	rec := httptest.NewRecorder()
	handleHealth(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.SHA != "0123456789abcdef0123456789abcdef01234567" {
		t.Errorf("expected injected sha, got %q", body.SHA)
	}
}
