package eventHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) FindEventsForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := contextutil.ValidateRoleInContext(ctx, models.BASIC); err != nil {
		logger.WarnContext(ctx, "get role from request", "error", err)
		serverutil.WriteResponse(w, handler.NewForbiddenErr(err), http.StatusBadRequest)
		return
	}
	userID, ok := ctx.Value(contextutil.UserIDKey).(string)
	if !ok {
		logger.ErrorContext(ctx, "user ID not found in context")
		serverutil.WriteResponse(w, errors.New("failed to get user ID from context").Error(), http.StatusInternalServerError)
		return
	}

	userEvents, err := a.Repos.Event.GetEventForUser(ctx, userID)
	if err != nil {
		switch {
		default:
			logger.ErrorContext(ctx, "failed to get the events for the user", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}
	if err := serverutil.Encode(w, http.StatusOK, userEvents); err != nil {
		logger.ErrorContext(ctx, "failed to send the user", "error", err)
		serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		return
	}
}
