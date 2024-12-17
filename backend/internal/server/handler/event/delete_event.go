package event

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) DeleteEvent() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get the event id from the url path
		eventID := r.PathValue("id")
		// use the service to delete that event
		resEventID, err := a.Svcs.Event.RemoveEvent(ctx, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to delete the event", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send the event id to the user to specify that the event is deleted properly.
		if err = serverutil.Encode(w, http.StatusInternalServerError, resEventID); err != nil {
			slog.ErrorContext(ctx, "failed to send the event ID", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
