package middleware

import (
	"context"
	"net/http"

	"math/rand"
)

type key int

const requestIDKey key = 42

// use the value in the context to log with the request ID key :
func AddRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := rand.Int63()
		ctx = context.WithValue(ctx, requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
