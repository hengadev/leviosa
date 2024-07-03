package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// to add a list of middleware together.
func Use(mws ...Middleware) http.Handler {
	var h http.Handler
	for _, mw := range mws {
		h = mw(h)
	}
	return h
}
