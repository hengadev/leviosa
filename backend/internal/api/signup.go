package api

import (
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
)

func (s *Server) signUpHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Handle the fact that admin can choose the role of a user
	switch r.Method {
	case http.MethodPost:
		user := getUserStoredFromRequest(w, r)
		if !user.ValidateEmail() || !user.ValidatePassword() {
			w.WriteHeader(http.StatusForbidden) // forbidden c'est pour quand on a pas les privileges pour realiser une action
			return
		}
		if s.Store.CheckUser(user.Email) { // sign up if and only if the user in not already registered.
			w.WriteHeader(http.StatusConflict)
			// TODO: Send in response body some message about the user already registered
			return
		}
		cookie, err := r.Cookie(types.SessionCookieName)
		if err != http.ErrNoCookie {
			print("J'ai un cookie\n")
			err := s.Store.CreateUser(user, s.Store.IsAdmin(cookie.Value))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			err := s.Store.CreateUser(user, false)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
		}
	default:
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *Server) createUser(user types.UserStored, cookie *http.Cookie) {

}
