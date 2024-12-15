package event

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (h *Handler) ModifyEvent() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		event, err := serverutil.Decode[eventService.Event](r.Body)
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode the event", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		eventID, err := h.Svcs.Event.ModifyEvent(ctx, &event)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update the event", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err = serverutil.Encode(w, http.StatusInternalServerError, eventID); err != nil {
			slog.ErrorContext(ctx, "failed to send the event ID", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
