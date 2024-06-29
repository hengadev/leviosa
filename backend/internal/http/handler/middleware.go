package handler

import (
	"fmt"
	"log"
	"log/slog"
	"time"
	// "log/slog"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
	"strings"
)

// NOTE: I get this thing from this parameter in the routes.go file
// I get it also from this video : https://www.youtube.com/watch?v=a2hWCvaAA80

// I think I should use the logger thing from the slog package.
type Logger struct {
	handler http.Handler
}

// from the samvcodes video
var loggerLevels = map[string]slog.Level{
	"info":  slog.LevelInfo,
	"debug": slog.LevelDebug,
	"error": slog.LevelError,
	"warn":  slog.LevelWarn,
}

// a very simple logger
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.handler.ServeHTTP(w, r)
	log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
}

// this function comes from the handler addes to all the endpoints in the NewServer constructor
func NewLogger(handlerToWrap http.Handler) *Logger {
	return &Logger{handlerToWrap}
	// NOTE: I get that from the page : https://drstearns.github.io/tutorials/gomiddleware/
}

// this function comes from the handler addes to all the endpoints in the NewServer constructor
func NewLoggingMiddleware(logger Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Just to test the logging thing")
	})
}

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

func Auth(next types.Handler, store Store) types.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		var cred struct {
			expectedRole types.Role
			sessionId    string
		}
		{
			header := r.Header.Get("Authorization")
			if cred.sessionId, cred.expectedRole = parseRequestAuth(r); cred.sessionId == header || cred.sessionId == "" {
				WriteResponse(w, "session ID invalid, you need to login again", http.StatusUnauthorized)
				return
			}
			if !store.HasSession(cred.sessionId) {
				WriteResponse(w, "Session has expired, you need to login again.", http.StatusUnauthorized)
				return
			}
			if !store.Authorize(cred.sessionId, cred.expectedRole) {
				WriteResponse(w, "The user does not have right to access this ressource.", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}
