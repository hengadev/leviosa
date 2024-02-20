package api

import (
	"encoding/json"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"log"
	"net/http"
)

func (s *Server) eventHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.showAllEvents(w)
	case http.MethodPost:
		s.makeEvent(r)
	case http.MethodDelete:
		s.deleteEvent(w, r)
	}
}

func (s *Server) showAllEvents(w http.ResponseWriter) {
	events := s.Store.GetAllEvents()
	if len(events) == 0 {
		w.WriteHeader(http.StatusNotFound)
	}
	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Fatal("Failed to encode events - ", err)
	}
}

func (s *Server) makeEvent(r *http.Request) {
	var event types.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Fatal("Failed to decode the body to get the event")
	}
	s.Store.PostEvent(&event)
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	event_id := r.URL.Query().Get("id")
	if !s.Store.CheckEvent(event_id) {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err := s.Store.DeleteEvent(event_id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}
