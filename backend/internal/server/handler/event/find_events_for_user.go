package event

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// func FindEventsForUser(eventRepo event.Reader) http.Handler {
func (h *Handler) FindEventsForUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		userID := ctx.Value(mw.UserIDKey).(int)
		resBody, err := h.Repos.Event.GetEventForUser(ctx, userID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the events for the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err := serverutil.Encode(w, http.StatusOK, resBody); err != nil {
			slog.ErrorContext(ctx, "failed to send the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
