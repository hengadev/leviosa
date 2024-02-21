package api

import (
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
)

func (s *Server) votesHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(types.SessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if s.Store.Authorize(cookie.Value, types.BASIC) {
		switch r.Method {
		case http.MethodPost:
			s.makeVote(w, r, cookie.Value)
		case http.MethodDelete:
			s.deleteVote(w, r)
		default:
			w.Header().Set("Access-Control-Allow-Methods", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func (s *Server) makeVote(w http.ResponseWriter, r *http.Request, session_id string) {
	event_id := r.URL.Query().Get("id")
	if !s.Store.CheckEvent(event_id) || event_id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user_id := s.Store.GetUserIdBySessionId(session_id)
	newVote := types.NewVote(&user_id, &event_id)
	if s.Store.CheckVote(&user_id, &event_id) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.Store.CreateVote(newVote); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := s.Store.DecreaseEventPlacecount(event_id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) deleteVote(w http.ResponseWriter, r *http.Request) {
	vote_id := r.URL.Query().Get("id")
	if vote_id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if s.Store.CheckVoteById(&vote_id) {
		if err := s.Store.DeleteVote(&vote_id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
