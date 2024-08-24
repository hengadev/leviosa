package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	app "github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
	"github.com/google/uuid"
)

type ContextKey int

const UserIDKey = ContextKey(23)

// Function middleware to authenticate and authorize users.
func Auth(s sessionService.Reader) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// make exception for certain path where you just call next.ServeHTTP(w,r)
			noAuthEndpoints := []string{
				serverutil.SIGNINENDPOINT,
				serverutil.SIGNUPENDPOINT,
			}
			url := strings.Join(strings.Split(r.URL.Path, "/")[3:], "/")
			for _, endpoint := range noAuthEndpoints {
				if url == endpoint {
					next.ServeHTTP(w, r)
					return
				}
			}
			ctx := r.Context()
			// get expected role from url path
			expectedRole := getExpectedRoleFromRequest(r)
			// get sessionID from request
			sessionID, err := getSessionIDFromRequest(r)
			if err != nil {
				slog.ErrorContext(ctx, "failed to get sessionID from the request", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// get session object from session repo
			session, err := s.FindSessionByID(ctx, sessionID)
			if err != nil {
				slog.ErrorContext(ctx, "failed to get the session object from sessionID", "error", err)
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			// validate session
			if pbms := session.Valid(ctx, expectedRole); len(pbms) > 0 {
				errors := "session error: ["
				for field, err := range pbms {
					errors += fmt.Sprintf("%s: %s, ", field, err)
				}
				errors += "]"
				slog.ErrorContext(ctx, "failed to validate session", "error", errors)
				http.Error(w, errors, http.StatusUnauthorized)
				return
			}
			// add userID to context.
			ctx = context.WithValue(ctx, UserIDKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getSessionIDFromRequest(r *http.Request) (string, error) {
	sessionID := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if err := uuid.Validate(sessionID); err != nil {
		return "", app.NewInvalidSessionErr(err)
	}
	return sessionID, nil
}

func getExpectedRoleFromRequest(r *http.Request) userService.Role {
	values := strings.Split(r.URL.Path, "/")
	return userService.ConvertToRole(values[3])
	// old api
	// var role user.Role
	// switch values[1] {
	// case user.ADMINISTRATOR.String():
	// 	role = user.ADMINISTRATOR
	// case user.GUEST.String():
	// 	role = user.GUEST
	// case user.BASIC.String():
	// 	role = user.BASIC
	// default:
	// 	role = user.UNKNOWN
	// }
	// return role
}
