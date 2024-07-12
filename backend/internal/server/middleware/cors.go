package middleware

import (
	"net/http"
	"strings"
)

// NOTE: find inspiration with this : https://github.com/rs/cors/blob/master/cors.go
// TODO: get that from an env variable
const (
	FRONTENDORIGIN = "http://localhost:5173"
)

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Find to do that things correctly
		(w).Header().Set("Access-Control-Allow-Origin", FRONTENDORIGIN)

		next.ServeHTTP(w, r)
	})
}

// TODO: Il me faut :
// - Une liste d'origine a open a l'application

// func EnableMethods(methods ...string) Middleware {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

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
