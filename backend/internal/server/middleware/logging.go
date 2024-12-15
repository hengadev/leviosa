package middleware

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/domainutil"
)

func AttachLogger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := rand.Int63()

			IP := r.Header.Get("X-Client-IP")
			if IP == "" {
				http.Error(w, "Cannot determine Client IP", http.StatusBadRequest)
				return
			}

			loggingSalt := os.Getenv("LOGGING_SALT")
			if loggingSalt == "" {
				http.Error(w, "Missing environment variable: LOGGING_SALT", http.StatusBadRequest)
				return
			}

			hashedIP := domainutil.HashWithSalt(IP, loggingSalt)

			logger = logger.With(
				"method", r.Method,
				"URL", r.URL.String(),
				"IP", hashedIP,
				"requestID", requestID,
			)

			ctx := context.WithValue(r.Context(), contextutil.LoggerKey, logger)

			logger.InfoContext(ctx, "Request started")

			next.ServeHTTP(w, r.WithContext(ctx))
			duration := time.Since(start)
			logger.InfoContext(ctx, "Request completed", "duration", duration)
		})
	}
}
