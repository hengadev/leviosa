package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthReportsStatusAndReleaseSHA(t *testing.T) {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	empty := ""
	injected := sha

	tests := []struct {
		name     string
		injected *string // nil: keep the default of a build without -ldflags -X
		want     string
	}{
		{name: "no SHA injected reports dev", injected: nil, want: "dev"},
		{name: "injected SHA is reported", injected: &injected, want: sha},
		{name: "empty injected SHA reports dev", injected: &empty, want: "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.injected != nil {
				// Simulate the -ldflags -X injection done at image build time.
				previous := ReleaseSHA
				ReleaseSHA = *tt.injected
				t.Cleanup(func() { ReleaseSHA = previous })
			}

			rec := httptest.NewRecorder()
			handleHealth(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("expected Content-Type application/json, got %q", ct)
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
			if body.SHA != tt.want {
				t.Errorf("expected sha %q, got %q", tt.want, body.SHA)
			}
		})
	}
}
