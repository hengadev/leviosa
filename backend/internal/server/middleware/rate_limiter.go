package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/GaryHY/event-reservation-app/pkg/serverutil"

	"golang.org/x/time/rate"
)

// Map that holds hashed IP address and their respective rate limiter using the token bucket algorithm.
var ipLimiters = map[string]*rate.Limiter{}

// PerIPRateLimit rates limit an endpoint at the rate of r attempt per ID per second. Client can allow up to b attempts at most.
func PerIPRateLimit(logger *slog.Logger, lim, burst int) Middleware {
	// TODO: use the logger, I want to get the information of users accesing rate limits in their use of the application
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: make sur that this part is pseudo anonymised so that I can make sure that this is safe in my database while protecting the app.
			// ie hash the IP address when stored
			// TODO: need a system to remove keys after a certain time
			ctx := r.Context()
			IP := getClientIP(r)
			if IP == "" || IP == "unknown" {
				logger.ErrorContext(ctx, "Invalid IP address")
				http.Error(w, "Invalid IP address", http.StatusBadRequest)
				return
			}
			// Check if the IP address is already present in the hash map
			limiter, exists := ipLimiters[IP]
			if !exists {
				// If the IP address is not present in the hash map, create a new rate limiter
				// limiter = rate.NewLimiter(2, 4)
				limiter = rate.NewLimiter(rate.Limit(lim), burst)
				ipLimiters[IP] = limiter
			}

			// Use the rate limiter to check if the request should be allowed
			if !limiter.Allow() {
				message := struct {
					Status string
					Body   string
				}{
					Status: "Request failed",
					Body:   "The API is at capacity, try again later",
				}
				if err := serverutil.Encode(w, http.StatusTooManyRequests, message); err != nil {
					// TODO: need to hash that IP address for the logs
					logger.ErrorContext(ctx, fmt.Sprintf("Rate limited %q for %s to %s", IP, r.Method, r.URL.String()))
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP retrieves the client IP set in the custom header 'X-Client-IP'. It returns an empty string if the IP address sent is invalid, unknown or could ne be parsed.
func getClientIP(r *http.Request) string {
	clientIP := r.Header.Get("X-Client-IP")
	if clientIP == "" || clientIP == "unknown" {
		return ""
	}
	if ip := net.ParseIP(clientIP); ip == nil {
		return ""
	}
	return clientIP
}
