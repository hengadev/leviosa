package api

import (
	"github.com/GaryHY/event-reservation-app/internal/types"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func (s *Server) signInHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	switch r.Method {
	case http.MethodOptions: // preflight request
		enableJSON(&w)
		enableMethods(&w, http.MethodPost)
	case http.MethodPost:
		user := getUserFromRequest(r)
		if types.IsNullGeneric(user) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !types.ValidateEmailGeneric(user) || !types.ValidatePasswordGeneric(user) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if !s.Store.CheckUser(user.Email) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		cookie, err := r.Cookie(types.SessionCookieName)
		if err == nil && s.Store.HasSession(cookie.Value) {
			w.WriteHeader(http.StatusOK)
			// TODO: Comment je gere la redirection vers la page d'accueil ?
			// redirectURL := fmt.Sprintf(os.Getenv("HOST"), "/home")
			// http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}
		hashpassword := s.Store.GetHashPassword(user)
		if err := bcrypt.CompareHashAndPassword([]byte(hashpassword), []byte(user.Password)); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user_id := s.Store.GetUserId(user.Email)
		session := types.NewSession(user_id)
		if err := s.Store.CreateSession(session); err != nil {
			log.Fatal("Failed to create session in the database for the user")
		}
		expired_at := session.Created_at.Add(types.SessionDuration)
		// TODO: Make the cookie secure to only use https using SSL encryption and use httpOnly.
		// How does that change the tests, and how to use https in golang ?
		http.SetCookie(w, &http.Cookie{
			Name:     types.SessionCookieName,
			Value:    session.Id,
			Expires:  expired_at,
			HttpOnly: true,
			Secure:   true,
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
