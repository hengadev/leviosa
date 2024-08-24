package middleware

import (
	"net/http"

	"github.com/GaryHY/event-reservation-app/pkg/serverutil"

	"golang.org/x/time/rate"
)

// NOTE: article : AAAAB3NzaC1yc2EAAAADAQABAAABgQDHCvLaXiYR8aMKctmwmtpBykB1JAXcNnK07ciHCKUbuWqgNRUT9M3g0zr8Jab+v8IjKLaa1manH+ynkKVVzJY+PbG7kpMUGSCfId3XHqC7WrmBvNNcUNcIcgptZf15ZBIpHD6YhDNDCmBZzM5b6XV+S3fGXVqZE0ARMrnbLEtyA8ZR+MD7S8xsiptJgZU3chNO+R/eLZwuqzbH96rjvhlnkqDcrGccihbOjnReBfpec7dgj4qy7knn5b4uh05LcCY0EXMtDI4MHN6Tv619FJ7PLv2gmtpflFP1ZiXWBUy/7nB20j2yTrPeDx0khxhLsPumAIdnPcdLCYzpx9+dlxRCvbv1G8KLFJ8MbqeurGHjERLHRQCTYJ7z0ALiDxB/Whh26yGAJmcjPBksovK3VQmXW9wDbSv3B5ZOPViMDhlUDHyanstUBhyzi0hgsu+kpWwhmMoMKjJBQaLoC5kxw/M7TQFUWzfai1B6TAtKrAHeuH9OlgAXj98nx4Q1YEfCKtE= henga@henga-KPL-W0X

func rateLimiter(next http.Handler) http.Handler {
	// TODO: find the right value for that.
	limiter := rate.NewLimiter(2, 4)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			message := struct {
				Status string
				Body   string
			}{
				Status: "Request failed",
				Body:   "The API is at capacity, try again later",
			}
			if err := serverutil.Encode(w, http.StatusTooManyRequests, message); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}
