package server

import (
	"fmt"
)

type ServerOption func(*Server)

func WithPort(port int) ServerOption {
	return func(s *Server) {
		addr := fmt.Sprintf(":%d", port)
		s.srv.Addr = addr
	}
}
