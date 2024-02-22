package api

import (
	"encoding/json"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"log"
	"net/http"
)

func (s *Server) adminEventHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(types.SessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if s.Store.Authorize(cookie.Value, types.ADMIN) {
		switch r.Method {
		case http.MethodGet:
			s.showAllEvents(w)
		case http.MethodPost:
			// TODO: Des qu'un event est cree je veux creer des cron jobs qui vont etre schedule en fonction de la date de planning de l'event en question.
			s.makeEvent(w, r)
		case http.MethodDelete:
			s.deleteEvent(w, r)
		case http.MethodPut:
			s.updateEvent(w, r)
		}
	}
}

func (s *Server) showAllEvents(w http.ResponseWriter) {
	events := s.Store.GetAllEvents()
	if len(events) == 0 {
		w.WriteHeader(http.StatusNotFound)
	}
	if err := json.NewEncoder(w).Encode(events); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Failed to encode events - ", err)
	}
}

func (s *Server) makeEvent(w http.ResponseWriter, r *http.Request) {
	var event types.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	event_id := r.URL.Query().Get("id")
	if !s.Store.CheckEvent(event_id) {
		w.WriteHeader(http.StatusBadRequest)
	}
	event := getEventFromRequest(w, r)
	if err := s.Store.UpdateEvent(&event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}
