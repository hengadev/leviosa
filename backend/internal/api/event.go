package api

import (
	"net/http"
)

func (s *Server) eventHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.showAllEvents(w)
	case http.MethodPost:
		s.makeEvent(r)
	}
}
