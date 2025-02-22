package eventHandler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

// handler
func (a *AppInstance) FindEventByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := contextutil.ValidateRoleInContext(ctx, models.ADMINISTRATOR); err != nil {
		logger.ErrorContext(ctx, "get role from request", "error", err)
		http.Error(w, handler.NewForbiddenErr(err), http.StatusBadRequest)
		return
	}
	eventID := r.PathValue("id")
	event, err := a.Repos.Event.GetEventByID(ctx, eventID)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrContext):
			logger.WarnContext(ctx, fmt.Sprintf("context error while trying to get event with ID %s", eventID))
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, rp.ErrNotFound):
			logger.WarnContext(ctx, "failed to find event with ID %q")
			http.Error(w, handler.NewNotFoundErr(err), http.StatusInternalServerError)
		case errors.Is(err, rp.ErrDatabase):
			logger.WarnContext(ctx, fmt.Sprintf("database query error while trying to get event with ID %s", eventID))
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}
	if err := serverutil.Encode(w, int(http.StatusOK), event); err != nil {
		logger.WarnContext(ctx, "failed to encode event ID for user")
		http.Error(w, fmt.Sprintf("Unable to encode event with ID of %q", eventID), http.StatusInternalServerError)
		return
	}
}
