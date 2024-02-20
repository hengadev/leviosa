package api

import (
	"github.com/GaryHY/event-reservation-app/internal/types"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func (s *Server) signInHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		user := getUserFromRequest(w, r)
		// TODO: Validate the mail and the password
		if !user.ValidateEmail() || !user.ValidatePassword() {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if !s.Store.CheckUser(user.Email) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// TODO: Put that in a get method when hitting the endpoint ?
		cookie, err := r.Cookie(types.SessionCookieName)
		if err == nil && s.Store.HasSession(cookie.Value) {
			w.WriteHeader(http.StatusOK)
			// TODO: Comment je gere la redirection vers la page d'accueil ?
			// http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
		hashpassword := s.Store.GetHashPassword(user)
		if err := bcrypt.CompareHashAndPassword([]byte(hashpassword), []byte(user.Password)); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// TODO: Get the id associated with the user mail
		user_id := s.Store.GetUserId(user.Email)
		session := types.NewSession(user_id)
		if err := s.Store.CreateSession(session); err != nil {
			log.Fatal("Failed to create session in the database for the user")
		}

		expired_at := session.Created_at.Add(types.SessionDuration)

		http.SetCookie(w, &http.Cookie{
			Name:    types.SessionCookieName,
			Value:   session.Id,
			Expires: expired_at,
		})

	default:
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
