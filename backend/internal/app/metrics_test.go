package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// scrape serves one GET /metrics from the registry's server handler.
func scrape(t *testing.T, srvHandler http.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	srvHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics: expected 200, got %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	return string(body)
}

func TestMetricsLabelRequestsByRoutePatternNeverByPath(t *testing.T) {
	reg, m := newMetricsRegistry()
	metricsHandler := newMetricsServer(reg).Handler

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("POST /contact", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	handler := m.middleware(mux, mux)

	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/users/jane.doe@example.com", nil),
		httptest.NewRequest(http.MethodPost, "/contact?email=jane.doe@example.com", nil),
		httptest.NewRequest(http.MethodGet, "/nowhere/secret-token", nil),
		httptest.NewRequest("PROPFIND", "/contact", nil),
	} {
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	body := scrape(t, metricsHandler)
	for _, want := range []string{
		`leviosa_http_requests_total{method="GET",route="GET /users/{user_id}",status="404"} 1`,
		`leviosa_http_requests_total{method="POST",route="POST /contact",status="200"} 1`,
		`leviosa_http_requests_total{method="GET",route="unmatched",status="404"} 1`,
		`method="OTHER"`,
		`leviosa_http_requests_in_flight 0`,
		`leviosa_build_info{sha="dev"} 1`,
		`go_goroutines`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output lacks %q", want)
		}
	}
	for _, leak := range []string{"jane.doe", "example.com", "secret-token", "PROPFIND"} {
		if strings.Contains(body, leak) {
			t.Errorf("metrics output contains %q from a request path or method", leak)
		}
	}
}

func TestStatusRecorderPassesFlushThrough(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}
	if err := http.NewResponseController(sr).Flush(); err != nil {
		t.Fatalf("flush through the recorder: %v", err)
	}
	if !rec.Flushed {
		t.Error("the underlying writer was not flushed")
	}
}
