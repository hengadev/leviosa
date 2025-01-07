package eventHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/event"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// TODO: add the fact that creating an event, should also create a vote and a table with the style votes_month_year.
func (a *AppInstance) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := contextutil.ValidateRoleInContext(ctx, models.ADMINISTRATOR); err != nil {
		logger.WarnContext(ctx, "get role from request", "error", err)
		serverutil.WriteResponse(w, errsrv.NewForbiddenErr(err), http.StatusBadRequest)
		return
	}
	event, err := serverutil.Decode[eventService.Event](r.Body)
	if err != nil {
		logger.WarnContext(ctx, "failed to decode event in create event handler", "error", err)
		serverutil.WriteResponse(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
		return
	}
	eventID, err := a.Svcs.Event.CreateEvent(ctx, &event)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrFormat):
			logger.WarnContext(ctx, "failed to format event for database operation")
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrNotCreated):
			logger.WarnContext(ctx, "failed to create event")
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, "database creating event failed")
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, rp.ErrContext):
			logger.WarnContext(ctx, "context error, deadline or timeout while creating event")
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrUnexpectedType):
			logger.WarnContext(ctx, "unexpected error while creating event")
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}
	if err := serverutil.Encode(w, http.StatusCreated, eventID); err != nil {
		logger.WarnContext(ctx, "failed to send the event", "error", err)
		serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		return
	}
}
