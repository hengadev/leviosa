package middleware

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	// "log/slog"
	"net/http"
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
