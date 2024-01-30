package api

import (
	"fmt"
	"net/http"
	"strings"
)

func NewServer(store Store) *Server {
	return &Server{Store: store}
}

type Server struct {
	Store Store
}

type Store interface {
	GetEventNameByID(id string) string
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/event/")
	name := s.Store.GetEventNameByID(id)
	if name == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprint(w, name)
}
