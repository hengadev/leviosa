package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/types"
)

func (s *Server) eventHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	cookie, err := r.Cookie(types.SessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if s.Store.Authorize(cookie.Value, types.BASIC) {
		switch r.Method {
		case http.MethodGet:
			id := r.URL.Query().Get("id")
			if id == "" {
				s.showAllEvents(w) // ne sera pas tester puisque deja pour le /admin/event endpoint
			} else {
				s.showUserEvents(w, id)
			}
		case http.MethodOptions:
			enableMethods(&w, http.MethodGet)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

// Retourne tous les events auxquels est inscrit un utilisateur donneee.
func (s *Server) showUserEvents(w http.ResponseWriter, id string) {
	if !s.Store.CheckUserById(id) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	events := s.Store.GetEventByUserId(id)
	if err := json.NewEncoder(w).Encode(&events); err != nil {
		log.Fatalf("Unable to encode the data for the events of the user identified by the id %q - %s", id, err)
	}
}
