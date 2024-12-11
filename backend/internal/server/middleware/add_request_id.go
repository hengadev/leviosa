package middleware

import (
	"context"
	"math/rand"
	"net/http"
)

type key int

const requestIDKey key = 42

// use the value in the context to log with the request ID key :
func AddRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := rand.Int63()
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
