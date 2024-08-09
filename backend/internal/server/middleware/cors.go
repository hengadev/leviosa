package middleware

import (
	"net/http"
	"os"
	"strings"
)

// NOTE: find inspiration with this : https://github.com/rs/cors/blob/master/cors.go

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// url := os.Getenv("HOST")
		url := os.Getenv("FRONTEND_ORIGIN")

		w.Header().Set("Access-Control-Allow-Origin", url)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		next.ServeHTTP(w, r)
	})
}

func EnableMethods(next http.Handler, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Methods", strings.Join(methods, " "))
		next.ServeHTTP(w, r)
	})
}

func EnableHeaders(next http.Handler, headers ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Headers", strings.Join(headers, " "))
		next.ServeHTTP(w, r)
	})
}

func EnableApplicationJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
