package event

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// handler
func (a *AppInstance) FindEventByID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		eventID := r.PathValue("id")
		event, err := a.Repos.Event.GetEventByID(ctx, eventID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the user ID", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err := serverutil.Encode(w, http.StatusOK, event); err != nil {
			slog.ErrorContext(ctx, "failed to encode the user ID", "error", err)
			http.Error(w, fmt.Sprintf("Unable to get event with the id of %q", eventID), http.StatusInternalServerError)
			return
		}
	})
}
