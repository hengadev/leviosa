package api

import (
	// "encoding/json"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"log"
	"net/http"
)

func (s *Server) photosHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(types.SessionCookieName)
	if err != nil {
		// w.WriteHeader(http.StatusUnauthorized)
		// return
	}
	switch r.Method {
	case http.MethodGet:
		event_id := r.URL.Query().Get("eventid")
		if s.Store.Authorize(cookie.Value, types.HELPER) {
			s.showAllPhotos(w, r, event_id)
		} else { // for the users
			s.showAllPhotosByUser(w, r, event_id)
		}
	case http.MethodPost:
		// if s.Store.Authorize(cookie.Value, types.HELPER) {
		s.postPhoto(w, r)
		// }
	case http.MethodDelete:
		s.deletePhoto(w, r)
	case http.MethodPut:
		if s.Store.Authorize(cookie.Value, types.ADMIN) {
			s.updatePhoto(w, r)
		}
	default:
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *Server) postPhoto(w http.ResponseWriter, r *http.Request) {
	event_id := r.URL.Query().Get("eventid")
	file, fileheader, err := r.FormFile("photo")
	if err != nil {
		log.Fatal("cannot read the file - ", err)
	}
	s.PhotoStore.PostFile(file, fileheader.Filename, event_id)
	w.Write([]byte("Le fichier a bien ete send !"))
}

func (s *Server) deletePhoto(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) updatePhoto(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) showAllPhotos(w http.ResponseWriter, r *http.Request, event_id string) {}

func (s *Server) showAllPhotosByUser(w http.ResponseWriter, r *http.Request, event_id string) {}
