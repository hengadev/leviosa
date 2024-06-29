package http

import (
	"net/http"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/http/service"
)

// NOTE: for a more in depth customization of the server I can consult this folder : /home/henga/Documents/projects/learn_golang/grpc/nicjackson/first
type Server struct {
	port              int
	host              string
	http.Handler      // the router of the application
	*handler.Services // contains all the services and repos
	// need to add this and see how to use it in the functions
	ReadTimeout       time.Time // 1s
	ReadHeaderTimeout time.Time
	WriteTimeout      time.Time
	IdleTimeout       time.Time
}

func NewServer(services *handler.Services, opts ...ServerOption) *Server {
	// build serve with default options.
	srv := &Server{
		port:     8000,
		host:     "localhost",
		Services: services,
	}
	// NOTE: the way I want my routers to be.
	// handlecreateAccount := srv.Svc.User.CreateAccount

	// add option for the server.
	for _, opt := range opts {
		opt(srv)
	}
	// create the router and add the routes
	router := &http.ServeMux{}
	// TODO: How to add all the services in here so that I can actually make all the routes since I need to make them easy brother.
	addRoutes(router, services)
	// add middlewares to all the routes.
	srv.Use(midw, midw)
	srv.Handler = router
	return srv
}

func midw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do the thing for that handler brother
		next.ServeHTTP(w, r)
	})
}

// I think that is what we need to do
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// addr := fmt.Sprintf(":%d", s.port)
	// slog.Info("HTTP server listening", "addr", addr)
	s.Handler.ServeHTTP(w, r)
}

type Middleware func(http.Handler) http.Handler

// A function to add middleware to all the routes of the multiplexer.
func (s *Server) Use(middlewares ...Middleware) {
	for _, mw := range middlewares {
		s.Handler = mw(s.Handler)
	}
}
