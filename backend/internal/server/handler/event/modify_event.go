package event

import (
	"log/slog"
	"net/http"

	eventService "github.com/GaryHY/event-reservation-app/internal/domain/event"
	errsrv "github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) ModifyEvent() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		event, err := serverutil.Decode[eventService.Event](r.Body)
		if err != nil {
			logger.ErrorContext(ctx, "failed to decode the event", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		err = a.Svcs.Event.ModifyEvent(ctx, &event)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update the event", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err = serverutil.Encode(w, http.StatusInternalServerError, event.ID); err != nil {
			logger.ErrorContext(ctx, "failed to send the event ID", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
