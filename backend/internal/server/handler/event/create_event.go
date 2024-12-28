package event

import (
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// TODO: add the fact that creating an event, should also create a vote and a table with the style votes_month_year.
func (a *AppInstance) CreateEvent() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// get the event from the body
		event, err := serverutil.Decode[eventService.Event](r.Body)
		if err != nil {
			logger.ErrorContext(ctx, "failed to decode the event", "error", err)
			serverutil.WriteResponse(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// use the service to make that event
		eventID, err := a.Svcs.Event.CreateEvent(ctx, &event)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create the event", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send the event id to the user
		if err := serverutil.Encode(w, http.StatusOK, eventID); err != nil {
			logger.ErrorContext(ctx, "failed to send the event", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
