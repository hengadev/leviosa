package photohandler

import (
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// func ShowAllPhotos(ph *photo.Service) http.Handler {
func (a *AppInstance) ShowAllPhotos() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		eventID := r.PathValue("id")
		// get the object from repository
		objects, err := a.Repos.Photo.GetAllObjects(ctx, eventID)
		if err != nil {

		}
		// send the object to the client
		if err := serverutil.Encode(w, http.StatusFound, objects); err != nil {
			logger.ErrorContext(ctx, "failed to send photos back to client", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
