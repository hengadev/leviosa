package http

type ServerOption func(*Server)

func WithPort(port int) ServerOption {
	return func(s *Server) {
		s.port = port
	}
}

func WithHost(address string) ServerOption {
	return func(s *Server) {
		s.host = address
	}
}
