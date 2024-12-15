package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/internal/server/service"
	// "github.com/GaryHY/event-reservation-app/pkg/config"
)

type Server struct {
	srv *http.Server
}

func New(handler *handler.Handler, logger *slog.Logger, opts ...ServerOption) *Server {
	// build server with default options.
	server := &Server{
		srv: &http.Server{
			// might need to change the values of the timeouts
			Addr:              ":5000",
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
	// TODO: I can  add the logger in here right ?
	// create router and add routes
	server.addRoutes(handler)
	// add middlewares common to all routes. [Order important]
	// TODO: remove this auth middleware here
	server.Use(
		mw.AttachLogger(logger),
		mw.Auth(handler.Repos.Session),
		// TODO: add that part later
		// mw.SetUserContext(handler.Repos.Session),
		mw.SetOrigin,
	)
	return server
}

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
