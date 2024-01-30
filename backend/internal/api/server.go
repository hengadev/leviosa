package api

import (
	"fmt"
	"net/http"
)

func NewServer(store Store) *Server {
	return &Server{Store: store}
}

type Server struct {
	Store Store
}

type Store interface {
	GetEventNameByID(id string) string
	PostEvent(name string)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.showEventBydID(w, r)
	case http.MethodPost:
		s.makeEvent(w, r)
	}
}

func (s *Server) showEventBydID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	name := s.Store.GetEventNameByID(id)
	if name == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprint(w, name)
}

func (s *Server) makeEvent(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	s.Store.PostEvent(name)
}
