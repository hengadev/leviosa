package api

import (
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
)

func (s *Server) votesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		makeVote(s, w, r)
	case http.MethodDelete:
		deleteVote(w, r)
	}
}

func makeVote(s *Server, w http.ResponseWriter, r *http.Request) {
	event_id := r.URL.Query().Get("id")
	if event_id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cookie, err := r.Cookie(types.SessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	session_id := cookie.Value
	user_id := s.Store.GetUserIdBySessionId(session_id)

	newVote := types.NewVote(&user_id, &event_id)
	if err := s.Store.CreateVote(newVote); err != nil {
		print("failed to create the vote")
	}
	if err = s.Store.DecreaseEventPlacecount(event_id); err != nil {
		print("failed to decrease the placount associated to the event")
	}

	w.WriteHeader(http.StatusCreated)
}

func deleteVote(w http.ResponseWriter, r *http.Request) {

}
