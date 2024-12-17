package photohandler

import (
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/app"
)

type AppInstance struct {
	*app.App
}

func (a *AppInstance) DeletePhoto() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func (a *AppInstance) UpdatePhoto() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func (a *AppInstance) ShowAllPhotosByUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

// func photosHandler(w http.ResponseWriter, r *http.Request) {
// 	// TODO: handle the authorization for helper or above using the session_id in the header authorization.
// 	enableCors(&w)
// 	cookie, err := r.Cookie(types.SessionCookieName)
// 	if err != nil {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		return
// 	}
// 	switch r.Method {
// 	case http.MethodGet:
// 		event_id := r.URL.Query().Get("eventid")
// 		if s.Store.Authorize(cookie.Value, types.HELPER) {
// 			s.showAllPhotos(w, r, event_id)
// 		}
// 		// } else { // for the users
// 		// 	s.showAllPhotosByUser(w, r, event_id)
// 		// }
// 	case http.MethodPost:
// 		// if s.Store.Authorize(cookie.Value, types.HELPER) {
// 		s.postPhoto(w, r)
// 		// }
// 	case http.MethodDelete:
// 		s.deletePhoto(w, r)
// 	case http.MethodPut:
// 		if s.Store.Authorize(cookie.Value, types.ADMIN) {
// 			s.updatePhoto(w, r)
// 		}
// 	case http.MethodOptions:
// 		enableMethods(&w, http.MethodPost)
// 	default:
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		return
// 	}
// }
