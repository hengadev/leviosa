package api

import (
	"encoding/json"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"log"
	"net/http"
)

func (s *Server) adminUsersHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	cookie, err := r.Cookie(types.SessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if s.Store.Authorize(cookie.Value, types.ADMIN) {
		switch r.Method {
		case "OPTIONS": // preflight request
			enableJSON(&w)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT")
		case http.MethodGet:
			s.showAllUsers(w)
		case http.MethodPost:
			s.signUpHandler(w, r)
		case http.MethodDelete:
			s.deleteUser(w, r)
		case http.MethodPut:
			s.updateUser(w, r)
		default:
			w.Header().Set("Access-Control-Allow-Methods", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func (s *Server) showAllUsers(w http.ResponseWriter) {
	users := s.Store.GetAllUsers()
	if len(users) == 0 {
		w.WriteHeader(http.StatusNotFound)
	}
	if err := json.NewEncoder(w).Encode(users); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Failed to encode events - ", err)
	}
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	user_id := r.URL.Query().Get("id")
	if !s.Store.CheckUserById(user_id) {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err := s.Store.DeleteUser(user_id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	user_id := r.URL.Query().Get("id")
	if !s.Store.CheckUserById(user_id) {
		w.WriteHeader(http.StatusBadRequest)
	}
	user := getUserStoredFromRequest(r)
	if types.IsNullGeneric(user) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.Store.UpdateUser(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}
