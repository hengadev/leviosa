package photohandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// func ShowAllPhotos(ph *photo.Service) http.Handler {
func (a *AppInstance) ShowAllPhotos() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		eventID := r.PathValue("id")
		// get the object from repository
		objects, err := a.Repos.Photo.GetAllObjects(ctx, eventID)
		if err != nil {

		}
		// send the object to the client
		if err := serverutil.Encode(w, http.StatusFound, objects); err != nil {
			slog.ErrorContext(ctx, "failed to send photos back to client", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
