package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
)

type sessionGetterFunc func(ctx context.Context, sessionID string) (*sessionService.Session, error)

// this is an authentication middleware, authorization is handle on a per route basis

// TODO:
// - handle what to do if there is nothing in the session because returning the function might not be ideal
// -> add non allowed endpoints

// get the role for the user in question
func SetUserContext(sessionGetter sessionGetterFunc) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()

			logger, ok := ctx.Value(contextutil.LoggerKey).(*slog.Logger)
			if !ok {
				http.Error(w, "logger not found in context", http.StatusInternalServerError)
				return
			}

			// get session ID from request
			sessionID, err := getSessionIDFromRequest(r)
			if err != nil {
				logger.ErrorContext(ctx, "get sessionID from request header", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// get session object from database
			session, err := sessionGetter(ctx, sessionID)
			if err != nil {
				logger.ErrorContext(ctx, "get session from database", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// set the role in the context using the session
			ctx = context.WithValue(r.Context(), contextutil.RoleKey, session.Role.String())
			ctx = context.WithValue(r.Context(), contextutil.UserIDKey, session.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
