package eventHandler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := contextutil.ValidateRoleInContext(ctx, models.ADMINISTRATOR); err != nil {
		logger.WarnContext(ctx, "get role from request", "error", err)
		http.Error(w, handler.NewForbiddenErr(err), http.StatusBadRequest)
		return
	}
	eventID := r.PathValue("id")
	err = a.Svcs.Event.RemoveEvent(ctx, eventID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotDeleted):
			logger.WarnContext(ctx, fmt.Sprintf("failed to delete event with ID %q", eventID))
			http.Error(w, handler.NewNotFoundErr(err), http.StatusInternalServerError)
		case errors.Is(err, rp.ErrContext):
			logger.WarnContext(ctx, fmt.Sprintf("context error while trying to delete event with ID %s", eventID))
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, fmt.Sprintf("database query error while trying to delete event with ID %s", eventID))
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrUnexpectedType):
			logger.WarnContext(ctx, fmt.Sprintf("unexpected error while trying to delete event with ID %s", eventID))
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}
	if err = serverutil.Encode(w, int(http.StatusNoContent), eventID); err != nil {
		logger.WarnContext(ctx, "failed to send event ID to user after deletion")
		http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		return
	}
}
