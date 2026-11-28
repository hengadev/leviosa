package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsAddr is where /metrics is served: a listener of its own, never the
// API port, so no proxy in front of the API (Caddy's /webhooks/*, the
// frontend) can ever expose it. Reached only over leviosa_network, by the
// Alloy agent in the production stack, which pushes to the homelab
// (infra/ansible/roles/observability).
const MetricsAddr = ":9090"

// Labels carry only the route pattern (`GET /rooms/{room_id}/metrics`),
// never the raw path, so no identifier, email or token ever reaches a metric:
// the metrics leave the hosting for the homelab, client data never does.
const unmatchedRoute = "unmatched"

// httpMetrics counts and times every request by route pattern.
type httpMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
}

func newHTTPMetrics(reg prometheus.Registerer) *httpMetrics {
	m := &httpMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "leviosa_http_requests_total",
			Help: "HTTP requests served, by method, route pattern and status code.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "leviosa_http_request_duration_seconds",
			Help:    "HTTP request duration, by method and route pattern.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"method", "route"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "leviosa_http_requests_in_flight",
			Help: "HTTP requests being served, open SSE streams included.",
		}),
	}
	reg.MustRegister(m.requests, m.duration, m.inFlight)
	return m
}

// knownMethods bounds the method label: anything else a client sends is
// counted as OTHER rather than becoming a new series.
var knownMethods = map[string]bool{
	http.MethodGet: true, http.MethodHead: true, http.MethodPost: true,
	http.MethodPut: true, http.MethodPatch: true, http.MethodDelete: true,
	http.MethodOptions: true,
}

// middleware wraps next (the whole chain), labelling each request with the
// pattern mux would route it to.
func (m *httpMetrics) middleware(mux *http.ServeMux, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, route := mux.Handler(r)
		if route == "" {
			route = unmatchedRoute
		}
		method := r.Method
		if !knownMethods[method] {
			method = "OTHER"
		}

		m.inFlight.Inc()
		defer m.inFlight.Dec()
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		m.requests.WithLabelValues(method, route, strconv.Itoa(rec.status)).Inc()
		m.duration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
	})
}

// statusRecorder keeps the status code; Flush and Unwrap pass through so SSE
// streams still flush.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	return r.ResponseWriter.Write(b)
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// newMetricsRegistry holds the Go runtime and process collectors, the
// release (leviosa_build_info{sha}) and the HTTP metrics.
func newMetricsRegistry() (*prometheus.Registry, *httpMetrics) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "leviosa_build_info",
		Help: "Always 1; the sha label is the release this backend runs.",
	}, []string{"sha"})
	buildInfo.WithLabelValues(releaseSHA()).Set(1)
	reg.MustRegister(buildInfo)
	return reg, newHTTPMetrics(reg)
}

// newMetricsServer serves reg at /metrics on MetricsAddr.
func newMetricsServer(reg *prometheus.Registry) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	return &http.Server{
		Addr:              MetricsAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
}

// runMetricsServer serves until shut down. A failure is logged, not fatal:
// the API keeps serving without its metrics (the homelab's "metrics missing"
// alert then fires).
func runMetricsServer(ctx context.Context, srv *http.Server, logger *slog.Logger) {
	logger.InfoContext(ctx, "Starting metrics server", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.ErrorContext(ctx, "metrics server stopped", "error", err)
	}
}
