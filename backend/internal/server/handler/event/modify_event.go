package eventHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/event"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) ModifyEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := contextutil.ValidateRoleInContext(ctx, models.ADMINISTRATOR); err != nil {
		logger.WarnContext(ctx, "get role from request", "error", err)
		serverutil.WriteResponse(w, handler.NewForbiddenErr(err), http.StatusBadRequest)
		return
	}
	event, err := serverutil.Decode[eventService.Event](r.Body)
	if err != nil {
		logger.WarnContext(ctx, "failed to decode the event", "error", err)
		serverutil.WriteResponse(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
		return
	}
	err = a.Svcs.Event.ModifyEvent(ctx, &event)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrContext):
			logger.WarnContext(ctx, "context error, deadline or timeout while updating event")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, "database updating event failed")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrNotUpdated):
			logger.WarnContext(ctx, "failed to update event")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrUnexpectedType):
			logger.WarnContext(ctx, "unexpected error while updating event")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrInvalidValue):
			logger.WarnContext(ctx, "invalid event error while updating event")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}
	if err = serverutil.Encode(w, http.StatusInternalServerError, event.ID); err != nil {
		logger.WarnContext(ctx, "failed to send event ID", "error", err)
		serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		return
	}
}
