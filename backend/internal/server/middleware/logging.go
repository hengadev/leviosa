package middleware

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/hengadev/leviosa/pkg/contextutil"
	"github.com/hengadev/leviosa/pkg/domainutil"
)

func AttachLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			start := time.Now()

			requestID := rand.Int63()

			// I just make a fake IP for now, I know my function to work:
			IP := "127.0.0.1"
			// TODO: use the getCgetClientIP function instead
			// IP := r.Header.Get("X-Client-IP")
			if IP == "" {
				slog.ErrorContext(ctx, "client IP not found with required header")
				http.Error(w, "Cannot determine Client IP", http.StatusBadRequest)
				return
			}

			loggingSalt := os.Getenv("LOGGING_SALT")
			if loggingSalt == "" {
				slog.ErrorContext(ctx, "logging salt not found in environment variables")
				http.Error(w, "Missing environment variable: LOGGING_SALT", http.StatusInternalServerError)
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
