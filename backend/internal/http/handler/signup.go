package handler

import (
	// "github.com/GaryHY/event-reservation-app/internal/mail"
	// "fmt"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
	// "os"
)

func (s *Server) signUpHandler(w http.ResponseWriter, r *http.Request) {
	// do the cors and all that thing in some middleware or it does not make any sense
	enableCors(&w)
	switch r.Method {
	case http.MethodOptions: // preflight request
		enableJSON(&w)
		enableMethods(&w, http.MethodPost)
	case http.MethodPost:
		cookie, err := r.Cookie(types.SessionCookieName)
		var user *types.UserStored
		switch err {
		case nil:
			if s.Store.Authorize(cookie.Value, types.ADMIN) {
				user = getUserStoredFromRequest(r)
			}
		case http.ErrNoCookie:
			userForm := getUserFormFromRequest(r)
			user = types.NewUserStored(userForm)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if types.IsNullGeneric(user) {
			w.WriteHeader(http.StatusBadRequest)
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
		// mail.SendWelcomeUserMail(user)
		w.WriteHeader(http.StatusCreated)
		// TODO: Comment je gere la redirection vers la page d'accueil ?
		// redirectURL := fmt.Sprintf(os.Getenv("HOST"), "/home")
		// http.Redirect(w, r, redirectURL, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
