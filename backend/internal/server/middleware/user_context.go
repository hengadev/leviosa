package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// TODO:
// - handle what to do if there is nothing in the session because returning the function might not be ideal
// -> add non allowed endpoints

// get the role for the user in question
func SetUserContext(sessionGetter sessionGetterFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			logger, ok := ctx.Value(contextutil.LoggerKey).(*slog.Logger)
			if !ok {
				slog.ErrorContext(ctx, "logger not found in context")
				serverutil.WriteResponse(w, "logger not found in context", http.StatusInternalServerError)
				return
			}

			// NOTE: that is just for dev mode
			exceptURL := []string{
				"hello",
				"user/register",
				"user/validate-otp",
				"user/approve-user",
				"oauth/google/user",
			}
			var url string
			url = strings.Join(strings.Split(r.URL.Path, "/")[3:], "/")
			for _, endpoint := range exceptURL {
				if url == endpoint {
					next.ServeHTTP(w, r)
					return
				}
			}

			// get session ID from request
			sessionID, err := getSessionIDFromRequest(r)
			if err != nil {
				logger.ErrorContext(ctx, "get sessionID from request header", "error", err)
				serverutil.WriteResponse(w, err.Error(), http.StatusBadRequest)
				return
			}

			// get session object from database
			session, err := sessionGetter(ctx, sessionID)
			if err != nil {
				logger.ErrorContext(ctx, "get session from database", "error", err)
				serverutil.WriteResponse(w, err.Error(), http.StatusBadRequest)
				return
			}

			// set the role in the context using the session
			ctx = context.WithValue(r.Context(), contextutil.RoleKey, session.Role.String())
			ctx = context.WithValue(r.Context(), contextutil.UserIDKey, session.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
