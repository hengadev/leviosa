package api

import (
	"encoding/json"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"log"
	"net/http"
)

func (s *Server) eventHandler(w http.ResponseWriter, r *http.Request) {
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
		default:
			w.Header().Set("Access-Control-Allow-Methods", "GET")
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
		log.Fatal("Unable to encode the data for the events of the user identified by the id %q - %s", id, err)
	}
}
