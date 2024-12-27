package event

import (
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) DeleteEvent() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// get the event id from the url path
		eventID := r.PathValue("id")
		// use the service to delete that event
		err = a.Svcs.Event.RemoveEvent(ctx, eventID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to delete the event", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send the event id to the user to specify that the event is deleted properly.
		if err = serverutil.Encode(w, http.StatusInternalServerError, eventID); err != nil {
			logger.ErrorContext(ctx, "failed to send the event ID", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
