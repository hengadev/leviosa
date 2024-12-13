package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

type contextKey string

const RoleKey = contextKey("role")

// get the role for the user in question
func SetRoleInContext(s sessionService.Reader) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get session id from request
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()

			sessionID, err := getSessionIDFromRequest(r)
			_ = sessionID
			if err != nil {
				slog.ErrorContext(ctx, "get sessionID from request", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// get role using the sessionID. Using the handler for that would have been great right
			// TODO: use the right function for that
			// role, err := s.GetRoleFromSessionID(sessionID)
			role := userService.ADMINISTRATOR

			ctx = context.WithValue(r.Context(), RoleKey, role.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
