package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/GaryHY/event-reservation-app/internal/api"
	"github.com/GaryHY/event-reservation-app/internal/types"

	// just for the test of using go routine
	"errors"
	"fmt"
	"time"
)

func parseRequestAuth(r *http.Request) (string, types.Role) {
	var role types.Role
	sessionId := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	values := strings.Split(r.URL.Path, "/")
	switch values[1] {
	case types.ADMIN.String():
		role = types.ADMIN
	case types.HELPER.String():
		role = types.HELPER
	default:
		role = types.BASIC
	}
	return sessionId, role
}

func Auth(next types.Handler, store api.Store) types.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		var cred struct {
			expectedRole types.Role
			sessionId    string
		}
		{
			header := r.Header.Get("Authorization")
			if cred.sessionId, cred.expectedRole = parseRequestAuth(r); cred.sessionId == header || cred.sessionId == "" {
				api.WriteResponse(w, "session ID invalid, you need to login again", http.StatusUnauthorized)
				return
			}
			if !store.HasSession(cred.sessionId) {
				api.WriteResponse(w, "Session has expired, you need to login again.", http.StatusUnauthorized)
				return
			}
			if !store.Authorize(cred.sessionId, cred.expectedRole) {
				api.WriteResponse(w, "The user does not have right to access this ressource.", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}

// TODO: For that function, I think that this link could help : https://github.com/luk4z7/go-concurrency-guide/blob/main/patterns/errorhandler/returnerror/main.go

// TODO:  try this one maybe ?
func otherAuth(next types.Handler, store api.Store) types.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		var cred struct {
			expectedRole types.Role
			sessionId    string
		}
		{
			header := r.Header.Get("Authorization")
			// TODO: change the api of parse, it should return an error for me to handle here.
			if cred.sessionId, cred.expectedRole = parseRequestAuth(r); cred.sessionId == header || cred.sessionId == "" {
				api.WriteResponse(w, "session ID invalid, you need to login again", http.StatusUnauthorized)
				return
			}
			// launch the go routine to speed up that process
			// TODO: make these functions return an error because it make more sense semantically.
			var errch = make(chan error)
			ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
			defer cancel()
			go userHasSession(ctx, cred.sessionId, errch)
			go userAuthorize(ctx, cred.sessionId, cred.expectedRole, errch)
			// should I use mutex for that or something
			select {
			case <-ctx.Done():
				// change the message with the one when the request is too long. That should be a constant
				http.Error(w, "", http.StatusInternalServerError)
				return
			case err := <-errch:
				switch err {
				case ErrNoValidSession:
					http.Error(w, fmt.Sprintln("%w", NewAuthError(ErrNoValidSession)), http.StatusInternalServerError)
					return
				case ErrNotAuthorized:
					http.Error(w, fmt.Sprintln("%w", NewAuthError(ErrNotAuthorized)), http.StatusInternalServerError)
					return
				}
			}
		}
		next(w, r)
	}
}

// Put that thing where I am defining the function to make the request to the DB
var (
	ErrNoValidSession = errors.New("no valid session")
	ErrNotAuthorized  = errors.New("not authorized")
)

// Put that part in api.go
var (
	ErrAuth = errors.New("invalid auth")
)

func NewAuthError(err error) error {
	return fmt.Errorf("%w: %w", ErrAuth, err)
}

// TODO: design these API perfectly
func userHasSession(ctx context.Context, sessionID string, errchan chan<- error) {
	time.Sleep(time.Second * 1)
	errchan <- errors.New("some error")
}

func userAuthorize(ctx context.Context, sessionId string, expectedRole types.Role, errchan chan<- error) {
	time.Sleep(time.Second * 3)
	errchan <- errors.New("some other error")
}
