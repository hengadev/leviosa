package api

import (
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
)

func (s *Server) signUpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		cookie, err := r.Cookie(types.SessionCookieName)
		var user *types.UserStored
		switch err {
		case nil:
			if s.Store.Authorize(cookie.Value, types.ADMIN) {
				user = getUserStoredFromRequest(w, r)
			}
		case http.ErrNoCookie:
			userForm := getUserFormFromRequest(w, r)
			user = types.NewUserStored(userForm)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !user.ValidateEmail() || !user.ValidatePassword() {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if s.Store.CheckUser(user.Email) {
			w.WriteHeader(http.StatusConflict)
			// TODO: Send in response body some message about the user already registered
			return
		}
		if err := s.Store.CreateUser(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)

	default:
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
