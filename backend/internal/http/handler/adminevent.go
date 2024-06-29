package handler

import (
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	// "os"

	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/stripe/stripe-go/v76"
)

func (s *Server) adminEventHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	cookie, err := r.Cookie(types.SessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if s.Store.Authorize(cookie.Value, types.ADMIN) {
		switch r.Method {
		case http.MethodOptions: // preflight request
			enableJSON(&w)
			enableMethods(&w, "*")
		// case http.MethodGet:
		// 	s.showAllEvents(w)
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

// func (s *Server) showAllEvents(w http.ResponseWriter) {
// 	events := s.Store.GetAllEvents()
// 	if len(events) == 0 {
// 		w.WriteHeader(http.StatusNotFound)
// 	}
// 	if err := json.NewEncoder(w).Encode(events); err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		log.Fatal("Failed to encode events - ", err)
// 	}
// }

func (s *Server) makeEvent(w http.ResponseWriter, r *http.Request) {
	var event types.EventForm
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Failed to decode the body to get the event - ", err)
	}
	defer r.Body.Close() // close the body after the reading of the struct from the request
	newevent := types.NewEvent(event.Location, event.PlaceCount, event.BeginAt, stripe.Price{}.ID)
	// TODO: to make that request, I need to be authorized and with https (https://docs.stripe.com/api/authentication)
	// priceid, err := createEventProductStripe(&w, newevent.Id, newevent.Date)
	// if err != nil {
	// 	log.Fatal("Failed to create the product for stripe - ", err)
	// }

	priceid := "" // to remove once I made the https connection
	newevent.PriceId = priceid
	s.Store.PostEvent(newevent)
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	event_id := r.URL.Query().Get("id")
	if !s.Store.CheckEvent(event_id) {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err := s.Store.DeleteEvent(event_id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	// TODO: Add that to the function when done with the function code
	// deleteEventProductStripe(event_id)
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
	// TODO: Make that function in payment to update the possible event, if that is somehting possible
	// updateEventProductStripe(event_id)
	w.WriteHeader(http.StatusCreated)
}
