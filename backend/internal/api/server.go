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
	id := strings.TrimPrefix(r.URL.Path, "/event/")
	name := s.Store.GetEventNameByID(id)
	if name == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprint(w, name)
}

func (s *Server) makeEvent(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/event/")
	fmt.Println("the name used is :", name)
	s.Store.PostEvent(name)
}
