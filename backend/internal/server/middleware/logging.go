package middleware

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/hengadev/leviosa/pkg/contextutil"
	"github.com/hengadev/leviosa/pkg/domainutil"
	mode "github.com/hengadev/leviosa/pkg/flags"
)

func AttachLogger(env mode.EnvMode, slogHandler slog.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			skipLogging := []string{
				"/healthz",
			}
			if slices.Contains(skipLogging, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			ctx := r.Context()

			requestID := rand.Int63()

			// I just make a fake IP for now, I know my function to work:
			var IP string
			switch env {
			case mode.ModeDev:
				IP = "127.0.0.1"
			case mode.ModeStaging, mode.ModeProd:
				// TODO: fix that one later with getting the right header for the IP address
				header := r.Header
				_ = header
				IP = "127.0.0.1"
			}

			if IP == "" {
				slog.ErrorContext(ctx, "client IP not found with required header")
				http.Error(w, "Cannot determine Client IP", http.StatusBadRequest)
				return
			}

			loggingSalt := os.Getenv("LOGGING_SALT")
			if loggingSalt == "" {
				slog.ErrorContext(ctx, "LOGGING_SALT not found in environment variables")
				http.Error(w, "Missing environment variable: LOGGING_SALT", http.StatusInternalServerError)
				return
			}

			hashedIP := domainutil.HashWithSalt(IP, loggingSalt)

			logger := slog.New(slogHandler)

			logger = logger.With(
				"method", r.Method,
				"URL", r.URL.String(),
				"IP", hashedIP,
				"requestID", requestID,
			)

			ctx = context.WithValue(r.Context(), contextutil.LoggerKey, logger)

			logger.InfoContext(ctx, "Request started")
			start := time.Now()
			next.ServeHTTP(w, r.WithContext(ctx))
			duration := time.Since(start)
			logger.InfoContext(ctx, "Request completed", "duration", duration)
		})
	}
}
