package event

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) FindEventsForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, ok := ctx.Value(contextutil.UserIDKey).(string)
	if !ok {
		logger.ErrorContext(ctx, "user ID not found in context")
		serverutil.WriteResponse(w, errors.New("failed to get user ID from context").Error(), http.StatusInternalServerError)
		return
	}

	resBody, err := a.Repos.Event.GetEventForUser(ctx, userID)
	if err != nil {
		switch {
		default:
			logger.ErrorContext(ctx, "failed to get the events for the user", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}
	if err := serverutil.Encode(w, http.StatusOK, resBody); err != nil {
		logger.ErrorContext(ctx, "failed to send the user", "error", err)
		serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		return
	}
}
