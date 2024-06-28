package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	// "time"

	"github.com/GaryHY/event-reservation-app/internal/types"
)

// TODO: I think I need to remove that since it is replaced by the  vote.
func (s *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	cookie, err := r.Cookie(types.SessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	eventId := r.URL.Query().Get("eventid")
	spot := r.URL.Query().Get("spot")
	switch r.Method {
	case http.MethodOptions:
		enableMethods(&w, http.MethodPost, http.MethodDelete)
	case http.MethodPost:
		if s.Store.Authorize(cookie.Value, types.BASIC) {
			spotInt, err := strconv.Atoi(spot)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				// TODO: Send some JSON error message or something to know what the error is
				return
			}
			s.makeRegistration(w, r, eventId, cookie.Value, spotInt)
		}
	case http.MethodDelete:
		if s.Store.Authorize(cookie.Value, types.ADMIN) {
			s.deleteRegistration(w, eventId, spot)
		}
	}
}

func (s *Server) makeRegistration(w http.ResponseWriter, r *http.Request, eventId, sessionId string, spot int) {
	// TODO:
	// - Get the price associated with the current eventid.
	// - Creer une table associe a l'event id en question.
	// - Pour chaque event j'aurais une session duration pour le creneau que j'envoie au front. A partir de cela je peux afficher tous les creneaux dispo a partir de le beginAt pour le creneau que j'envoie au front. A partir de cela je peux afficher tous les creneaux dispo a partir de le beginAt. Et pour cette requete, je vais juste envoyer un int qui correspond a la numerotation du creneau et avec cela je peux savoir quand il commence.

	if !s.Store.CheckEvent(eventId) || eventId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userId, err := s.Store.GetUserIdBySessionId(sessionId)
	if err != nil {
		WriteResponse(w, "Unable to get the events from the database", http.StatusInternalServerError)
		return
	}
	//partie ou je check si il n'y a pas de reservation sur ce creneau
	eventBeginAt := s.Store.GetBeginAtByEventId(eventId) // get  that part from the eventId
	registrationBeginAt := eventBeginAt.Add(types.RegistrationDuration)
	registration := types.NewRegistration(userId, eventId, registrationBeginAt)
	if s.Store.CheckRegistration(registration) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// partie ou je fais la reservation en ajoutant les infos qu'il faut dans la table associee a l'event
	if err := s.Store.CreateRegistration(registration); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// checkout pour payer
	priceId := s.Store.GetPriceIDByEventID(eventId)
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("STRIPE_SECRET_KEY")))
	if err := createCheckoutSession(w, priceId); err != nil {
		// TODO: Do better error handling based on the different errors that you can get
		println("Failed to create session for the user")
	}

	// TODO: How to handle that webhook thing for the payment thing

	// je reduis le nombre de places pour l'event en question
	if err := s.Store.DecreaseEventPlacecount(eventId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// NOTE: do you want to do that ? You would have to give people their money back.
func (s *Server) deleteRegistration(w http.ResponseWriter, eventId, sessionId string) {
}
