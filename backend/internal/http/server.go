package http

import (
	"context"
	"net/http"
	"time"

	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/internal/http/service"
)

type Server struct {
	srv *http.Server
}

func NewServer(handler *handler.Handler, opts ...ServerOption) *Server {
	// build server with default options.
	server := &Server{
		srv: &http.Server{
			// might need to change these values
			IdleTimeout:       120 * time.Second,
			ReadTimeout:       1 * time.Second,
			WriteTimeout:      1 * time.Second,
			ReadHeaderTimeout: 1 * time.Second,
		},
	}
	// add server options
	for _, opt := range opts {
		opt(server)
	}
	// create router and add routes
	server.addRoutes(handler)
	// add middlewares common to all routes. [Order important]
	server.Use(
		mw.AddRequestID,
		mw.Auth(handler.Repos.Session),
		mw.Cors,
		mw.EnableApplicationJSON,
		mw.TestPrint,
	)
	return server
}

// I think that is what we need to do
func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	return s.srv.ListenAndServeTLS(certFile, keyFile)
}

func (s *Server) Close() error {
	return s.srv.Close()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

// A function to add middleware to all the routes of the multiplexer.
func (s *Server) Use(middlewares ...mw.Middleware) {
	for _, mw := range middlewares {
		s.srv.Handler = mw(s.srv.Handler)
	}
}
