package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func RequestLogging(logger *slog.Logger) Middleware {
	// TODO: old format for logging
	// logger.InfoContext(ctx, "[INFO] Request started - Method: %s, URL: %s, IP: %s, requestID: %s", r.Method, r.URL.String(), r.RemoteAddr, requestID)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()
			requestID, ok := ctx.Value(requestIDKey).(int64)
			if !ok {
				logger.ErrorContext(ctx, "no request ID found for the logger")
			}
			logger = logger.With(
				"method", r.Method,
				"URL", r.URL.String(),
				// TODO: make sure to use the right IP address that should be found in some headers depending on the provider that I use
				"IP", r.RemoteAddr,
				"requestID", requestID,
			)
			logger.InfoContext(ctx, "Request started")
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			logger.InfoContext(ctx, "Request completed", "duration", duration)
		})
	}
}
