package api

import (
	"encoding/json"
	// "fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
)

func NewServer(store Store) *Server {
	server := new(Server)
	server.Store = store

	router := http.NewServeMux()
	router.Handle("/users", http.HandlerFunc(server.usersHandler))
	router.Handle("/event", http.HandlerFunc(server.eventHandler))
	server.Handler = router

	return server
}

type Server struct {
	Store Store
	http.Handler
}

type Store interface {
	GetEventNameByID(id string) types.Event
	PostEvent(name string)

	// TODO: Model that using the Event type implemented
	// GetEventNameByID(id string) Event
	// PostEvent(event Event)
}

func (s *Server) usersHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) eventHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.showEventBydID(w, r)
	case http.MethodPost:
		s.makeEvent(r)
	}
}

func (s *Server) showEventBydID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	event := s.Store.GetEventNameByID(id)
	if event.Name == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	json.NewEncoder(w).Encode(&event)
}

func (s *Server) makeEvent(r *http.Request) {
	name := r.URL.Query().Get("name")
	s.Store.PostEvent(name)
}
