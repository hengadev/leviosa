package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// change the store here to a session store to use the interface for that.
func Auth(store *session.Reader) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var cred struct {
				expectedRole user.Role
				sessionId    string
			}
			{
				header := r.Header.Get("Authorization")
				if cred.sessionId, cred.expectedRole = parseRequestAuth(r); cred.sessionId == header || cred.sessionId == "" {
					serverutil.WriteResponse(w, "session ID invalid, you need to login again", http.StatusUnauthorized)
					return
				}
				// TODO: Do i need the next two conditions split ?
				if !store.HasSession(cred.sessionId) {
					serverutil.WriteResponse(w, "Session has expired, you need to login again.", http.StatusUnauthorized)
					return
				}
				if !store.Authorize(cred.sessionId, cred.expectedRole) {
					serverutil.WriteResponse(w, "The user does not have right to access this ressource.", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// use that and the logger to log the error if this is something that I want.
var (
	ErrNoValidSession = errors.New("no valid session")
	ErrNotAuthorized  = errors.New("not authorized")
)

func parseRequestAuth(r *http.Request) (string, user.Role) {
	var role user.Role
	sessionId := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	values := strings.Split(r.URL.Path, "/")
	switch values[1] {
	case user.ADMINISTRATOR.String():
		role = user.ADMINISTRATOR
	case user.GUEST.String():
		role = user.GUEST
	case user.BASIC.String():
		role = user.BASIC
	default: // the default is an error since the role is not known
		role = user.UNKNOWN
	}
	return sessionId, role
}
