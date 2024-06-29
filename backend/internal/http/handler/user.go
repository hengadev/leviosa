package handler

import (
	"encoding/json"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/types"
)

func (s *Server) userHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	sessionId := parseAuthorizationHeader(r)

	if sessionId == "" {
		WriteResponse(w, "No session id in the request, you need to login again", http.StatusUnauthorized)
		return
	}
	if sessionId != "" && !s.Store.HasSession(sessionId) {
		WriteResponse(w, "Session has expired, you need to login again.", http.StatusUnauthorized)
		return
	}
	if s.Store.Authorize(sessionId, types.BASIC) {
		switch r.Method {
		case http.MethodOptions:
			enableMethods(&w, http.MethodGet)
		case http.MethodGet:
			user := s.Store.GetUserFromSessionId(sessionId)
			w.Header().Set("Content-Type", "application/json")
			// TODO: should I use a pointer instead ?
			if err := json.NewEncoder(w).Encode(user); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}
