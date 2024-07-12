package middleware

import (
	"fmt"
	"net/http"
)

func TestPrint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := ctx.Value(requestIDKey).(int64)
		fmt.Fprintln(w, "test middleware :", id)
		next.ServeHTTP(w, r)
	})
}
