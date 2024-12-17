package photohandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// func PostPhoto(ph *photo.Service) http.Handler {
func (a *AppInstance) PostPhoto() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		eventID := r.PathValue("id")
		file, fileheader, err := r.FormFile("photo")
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the photo information from form", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// post file to bucket
		url, err := a.Svcs.Photo.PostFile(ctx, file, fileheader.Filename, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to post file", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
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
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
