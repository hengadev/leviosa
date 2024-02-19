package api

import (
	"encoding/json"
	// "fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"log"
	"net/http"
)

// TODO: Use json to get the data

func (s *Server) showEventBydID(w http.ResponseWriter, r *http.Request) {
	// TODO: Use that to get the event
	// var event types.Event
	// json.NewDecoder(r.Body).Decode(event)

	id := r.URL.Query().Get("id")
	event := s.Store.GetEventByID(id)
	if event.Id == "" {
		w.WriteHeader(http.StatusNotFound)
	}

	json.NewEncoder(w).Encode(&event)
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

func getUserFromRequest(w http.ResponseWriter, r *http.Request) (user *types.User) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return &types.User{}
	}
	return
}

func getUserStoredFromRequest(w http.ResponseWriter, r *http.Request) (user *types.UserStored) {
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return &types.UserStored{}
	}
	return
}

func getEventFromRequest(w http.ResponseWriter, r *http.Request) (event types.Event) {
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return types.Event{}
	}
	return
}
