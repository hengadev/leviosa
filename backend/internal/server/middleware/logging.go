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
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func AttachLogger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			start := time.Now()

			requestID := rand.Int63()

			// I just make a fake IP for now, I know my function to work:
			IP := "127.0.0.1"
			// IP := r.Header.Get("X-Client-IP")
			if IP == "" {
				slog.ErrorContext(ctx, "client IP not found with required header")
				serverutil.WriteResponse(w, "Cannot determine Client IP", http.StatusBadRequest)
				return
			}

			loggingSalt := os.Getenv("LOGGING_SALT")
			if loggingSalt == "" {
				slog.ErrorContext(ctx, "logging salt not found in environment variables")
				serverutil.WriteResponse(w, "Missing environment variable: LOGGING_SALT", http.StatusInternalServerError)
				return
			}

			hashedIP := domainutil.HashWithSalt(IP, loggingSalt)

			logger = logger.With(
				"method", r.Method,
				"URL", r.URL.String(),
				"IP", hashedIP,
				"requestID", requestID,
			)

			ctx = context.WithValue(r.Context(), contextutil.LoggerKey, logger)

			logger.InfoContext(ctx, "Request started")

			next.ServeHTTP(w, r.WithContext(ctx))
			duration := time.Since(start)
			logger.InfoContext(ctx, "Request completed", "duration", duration)
		})
	}
}
