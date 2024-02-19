package api

import "net/http"

func (s *Server) eventByIdHandler(w http.ResponseWriter, r *http.Request) {
	// ici que des methodes post et get
	switch r.Method {
	case http.MethodGet:
		s.showEventBydID(w, r)
	case http.MethodPost:
	default:
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
