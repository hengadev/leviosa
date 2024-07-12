package photohandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/photo"
	"github.com/GaryHY/event-reservation-app/internal/http/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func PostPhoto(ph *photo.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		eventID := r.PathValue("id")
		file, fileheader, err := r.FormFile("photo")
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the photo information from form", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// post file to bucket
		url, err := ph.PostFile(ctx, file, fileheader.Filename, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to post file", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// TODO: see if I can send the url without making an object.
		// send the url back to the user
		// type Response struct {
		// 	URL string `json:"url"`
		// }
		// if err := serverutil.Encode(w, http.StatusSeeOther, Response{URL: url}); err != nil {
		if err := serverutil.Encode(w, http.StatusSeeOther, url); err != nil {
			slog.ErrorContext(ctx, "failed to send url back to client", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func ShowAllPhotos(ph *photo.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		eventID := r.PathValue("id")
		// get the object from repository
		objects, err := ph.GetAllObjects(ctx, eventID)
		if err != nil {

		}
		// send the object to the client
		if err := serverutil.Encode(w, http.StatusFound, objects); err != nil {
			slog.ErrorContext(ctx, "failed to send photos back to client", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func DeletePhoto() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func UpdatePhoto() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func ShowAllPhotosByUser() http.Handler {
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
