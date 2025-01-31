package middleware

import (
	"net/http"
	"os"
	"strings"
)

// NOTE: find inspiration with this : https://github.com/rs/cors/blob/master/cors.go

// NOTE: the headers that I need to set
// w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
// w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
// w.Header().Set("Access-Control-Allow-Credentials", "true")

func SetOrigin(next Handlerfunc) Handlerfunc {
	allowedOrigins := map[string]bool{
		os.Getenv("FRONTEND_ORIGIN"): true,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		next(w, r)
	}
}

func EnableMethods(next http.Handler, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, " "))
		next.ServeHTTP(w, r)
	})
}

func EnableHeaders(next http.Handler, headers ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, " "))
		next.ServeHTTP(w, r)
	})
}

func EnableCredentials(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}

func EnableApplicationJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
