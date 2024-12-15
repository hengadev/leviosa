package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
)

// this is an authentication middleware, authorization is handle on a per route basis

// get the role for the user in question
func SetUserContext(s sessionService.Reader) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get session id from request
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()

			logger, ok := ctx.Value(contextutil.LoggerKey).(*slog.Logger)
			if !ok {
				http.Error(w, "logger not found in context", http.StatusInternalServerError)
				return
			}

			sessionID, err := getSessionIDFromRequest(r)
			if err != nil {
				logger.ErrorContext(ctx, "get sessionID from request header", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			session, err := s.FindSessionByID(ctx, sessionID)
			if err != nil {
				// do better error handling with the right error value
				logger.ErrorContext(ctx, "get sessionID from request", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// set the role in the context using the session
			ctx = context.WithValue(r.Context(), contextutil.RoleKey, session.Role.String())
			// TODO:set the userID in the context

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
