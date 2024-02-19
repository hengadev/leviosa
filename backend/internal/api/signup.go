package api

import (
	"net/http"
)

func (s *Server) signUpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		user := getUserFromRequest(w, r)
		if !user.ValidateEmail() || !user.ValidatePassword() {
			w.WriteHeader(http.StatusForbidden) // forbidden c'est pour quand on a pas les privileges pour realiser une action
			return
		}
		if s.Store.CheckUser(user.Email) { // sign up if and only if the user in not already registered.
			w.WriteHeader(http.StatusConflict)
			// TODO: Send in response body some message about the user already registered
			return
		}
		if err := s.Store.CreateUser(user); err == nil {
			w.WriteHeader(http.StatusCreated)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
